package main

import (
	"fmt"
	"github.com/fhs/gompd/mpd"
	"github.com/golang/glog"
	"strconv"
	"strings"
	"time"
)

func mpdConnect() *mpd.Client {
	con, err := mpd.Dial("tcp", configuration.MPDConf.Host)
	if err != nil {
		glog.Error("mpdConnect error: ", err.Error())
	}
	return con
}

func connClose(con *mpd.Client) {
	if con != nil {
		con.Close()
	}
}

// MPDStatus of MPD daemon
type MPDStatus struct {
	CurrentSong string
	Playing     bool
	Status      string
	Flags       string
	Volume      string
	Error       string
}

func (s *MPDStatus) String() string {
	return fmt.Sprintf("MPDStatus[Playing=%v Status=%v Flags=%v Volume=%v CurrentSong='%v']",
		s.Playing, s.Status, s.Flags, s.Volume, s.CurrentSong)
}

// MPD client
type MPD struct {
	Message chan *MPDStatus
	watcher *mpd.Watcher
	end     chan bool
	active  bool
}

// NewMPD create new MPD client
func NewMPD() *MPD {
	return &MPD{
		Message: make(chan *MPDStatus, 5),
		end:     make(chan bool),
		active:  true,
	}
}

func (m *MPD) watch() (err error) {
	m.watcher, err = mpd.NewWatcher("tcp", configuration.MPDConf.Host, "")
	if err != nil {
		glog.Errorf("mpd.watch: connect to %v error: %v", configuration.MPDConf.Host, err.Error())
		return err
	}

	glog.Info("mpd.watch: connected to ", configuration.MPDConf.Host)
	glog.V(1).Infof("mpd.watch: starting watch")

	m.Message <- MPDGetStatus()

	for {
		if m.watcher == nil {
			break
		}
		select {
		case _ = <-m.end:
			glog.Info("mpd.watch: end")
			m.active = false
			return
		case subsystem := <-m.watcher.Event:
			if glog.V(1) {
				glog.Info("mpd.watch: event: ", subsystem)
			}
			m.Message <- MPDGetStatus()
			/*
				switch subsystem {
				case "player":
					m.Message <- MPDGetStatus()
				default:
					m.Message <- MPDGetStatus()
				}
			*/
		case err := <-m.watcher.Error:
			//glog.Errorf("mpd.watch: error event: %v", err)
			return err
		}
	}

	return nil
}

// Connect to mpd daemon
func (m *MPD) Connect() (err error) {
	go func() {
		defer func() {
			glog.Infof("mpd.watch: closing")
			if m.watcher != nil {
				m.watcher.Close()
				m.watcher = nil
				close(m.end)
			}
		}()

		for m.active {
			if err = m.watch(); err != nil {
				glog.Errorf("mpd.Connect: start watch error: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	return
}

// Close MPD client
func (m *MPD) Close() {
	glog.V(1).Info("mpd.Close")
	if m.watcher != nil {
		m.end <- true
	}
}

// GetStatus connect to mpd and get current status
func MPDGetStatus() (s *MPDStatus) {
	s = &MPDStatus{
		Flags: "ERR",
	}

	con := mpdConnect()
	if con == nil {
		return
	}
	if glog.V(1) {
		glog.Info("mpd.GetStatus: connected to ", configuration.MPDConf.Host)
	}

	defer connClose(con)

	status, err := con.Status()
	if err != nil {
		glog.Errorf("mpd.GetStatus: Status error: %v", err.Error())
		return
	}

	s.Status = status["state"]
	s.Playing = s.Status != "stop"
	if s.Playing {
		s.Flags = ""
	} else {
		s.Error = status["error"]
	}
	s.Volume = status["volume"]

	if status["random"] != "0" {
		s.Flags = "S"
	} else if status["repeat"] != "0" {
		s.Flags = "R"
	}

	song, err := con.CurrentSong()
	if err != nil {
		glog.Errorf("mpd.GetStatus: CurrentSong error: %v", err.Error())
		return
	}

	//glog.Infof("Status: %+v", status)
	//glog.Infof("Song: %+v", song)

	var res []string

	currsongnum := ""
	if a, ok := status["song"]; ok {
		if song, err := strconv.Atoi(a); err == nil {
			currsongnum = strconv.Itoa(song + 1)
		}
	}
	if a, ok := status["playlistlength"]; ok {
		currsongnum += "/" + a
	}
	if currsongnum != "" {
		res = append(res, currsongnum)
	}

	hasATN := false

	add := func(key string, atn bool) {
		if a, ok := song[key]; ok && a != "" {
			res = append(res, a)
			hasATN = atn || hasATN
		}
	}

	add("Name", true)
	add("Artist", true)
	add("Title", true)
	add("Track", false)
	add("Album", false)

	if !hasATN {
		add("file", false)
	}

	s.CurrentSong = strings.Join(res, "; ")
	return
}

func MPDPlay(index int) {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		con.Play(index)
	}
}

func MPDStop() {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		con.Stop()
	}
}

func MPDPause() {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		if stat, err := con.Status(); err == nil {
			err = con.Pause(stat["state"] != "pause")
		}
	}
}

func MPDNext() {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		con.Next()
	}
}

func MPDPrev() {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		con.Previous()
	}
}

func changeVol(change int) {
	con := mpdConnect()
	if con == nil {
		return
	}
	defer connClose(con)
	if stat, err := con.Status(); err == nil {
		vol, err := strconv.Atoi(stat["volume"])
		if err != nil {
			return
		}
		vol += change
		if vol > 100 {
			vol = 100
		} else if vol < 0 {
			vol = 0
		}
		con.SetVolume(vol)
	}
}

func MPDVolUp() {
	changeVol(5)
}

func MPDVolDown() {
	changeVol(-5)
}

func MPDPlaylists() (pls []string) {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		playlists, err := con.ListPlaylists()
		if err == nil {
			for _, pl := range playlists {
				pls = append(pls, pl["playlist"])
			}
		} else {
			glog.Error("MPD.Playlists list error: ", err)
		}
	}
	return
}

func MPDPlayPlaylist(playlist string) {
	con := mpdConnect()
	if con != nil {
		defer connClose(con)
		con.Clear()
		con.PlaylistLoad(playlist, -1, -1)
		con.Play(0)
	}
}

func MPDCurrPlaylist() (pls []string, pos int) {
	con := mpdConnect()
	if con == nil {
		return
	}
	defer connClose(con)
	if stat, err := con.Status(); err == nil {
		pos, _ = strconv.Atoi(stat["song"])
	}
	playlists, err := con.PlaylistInfo(-1, -1)
	if err != nil {
		glog.Error("MPD.CurrPlaylist list error: ", err)
		return
	}
	for _, pl := range playlists {
		if title, ok := pl["Title"]; ok {
			pls = append(pls, title)
		} else {
			pls = append(pls, pl["file"])
		}
	}
	return
}

var preMuteVol = -1

func MPDVolMute() {
	con := mpdConnect()
	if con == nil {
		return
	}
	defer connClose(con)

	stat, err := con.Status()
	if err != nil {
		glog.Error("MPD.MPDVolMute error: ", err)
		return
	}

	vol, err := strconv.Atoi(stat["volume"])
	if err != nil {
		return
	}

	if vol == 0 {
		if preMuteVol > 0 {
			con.SetVolume(preMuteVol)
		} else {
			con.SetVolume(100)
		}
	} else {
		preMuteVol = vol
		con.SetVolume(0)
	}
}
