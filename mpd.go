package main

import (
	"flag"
	"fmt"
	"github.com/fhs/gompd/mpd"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	mpdHost = flag.String("mpdHost", "pi:6600", "MPD address")
)

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
		Message: make(chan *MPDStatus, 10),
		end:     make(chan bool, 1),
		active:  true,
	}
}

func (m *MPD) watch() (err error) {
	m.watcher, err = mpd.NewWatcher("tcp", *mpdHost, "")
	if err != nil {
		log.Printf("mpd.watch: connect to %v error: %v", *mpdHost, err.Error())
		return err
	}
	log.Printf("mpd.watch: connected to %v", *mpdHost)

	log.Println("mpd.watch: starting watch")

	m.Message <- MPDGetStatus()

	for {
		if m.watcher == nil {
			break
		}
		select {
		case _ = <-m.end:
			log.Printf("mpd.watch: end")
			return
		case subsystem := <-m.watcher.Event:
			log.Printf("mpd.watch: event: %v", subsystem)
			switch subsystem {
			case "player":
				m.Message <- MPDGetStatus()
			default:
				m.Message <- MPDGetStatus()
			}
		case err := <-m.watcher.Error:
			log.Printf("mpd.watch: error event: %v", err)
			return err
		}
	}

	return nil
}

// Connect to mpd daemon
func (m *MPD) Connect() (err error) {

	go func() {
		defer func() {
			log.Println("mpd.watch: closing")
			if m.watcher != nil {
				m.watcher.Close()
				m.watcher = nil
				close(m.end)
			}
		}()

		for m.active {
			if err = m.watch(); err != nil {
				log.Printf("mpd.Connect: start watch error: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	return
}

// Close MPD client
func (m *MPD) Close() {
	log.Printf("mpd.Close")
	m.active = false
	if m.watcher != nil {
		m.end <- true
	}
}

// GetStatus connect to mpd and get current status
func MPDGetStatus() (s *MPDStatus) {
	s = &MPDStatus{
		Playing:     false,
		Flags:       "ERR",
		CurrentSong: "",
	}

	con, err := mpd.Dial("tcp", *mpdHost)
	if err != nil {
		log.Printf("mpd.GetStatus: connect do %s error: %v", *mpdHost, err.Error())
		return
	}
	log.Printf("mpd.GetStatus: connected to %s", *mpdHost)

	defer con.Close()

	status, err := con.Status()
	if err != nil {
		log.Printf("mpd.GetStatus: Status error: %v", err.Error())
		return
	}

	s.Status = status["state"]
	s.Error = status["error"]
	s.Playing = s.Status != "stop"
	s.Volume = status["volume"]

	if status["random"] != "0" {
		s.Flags = "S"
	} else if status["repeat"] != "0" {
		s.Flags = "R"
	}

	song, err := con.CurrentSong()
	if err != nil {
		log.Printf("mpd.GetStatus: CurrentSong error: %v", err.Error())
		return
	}

	//log.Printf("Status: %+v", status)
	//log.Printf("Song: %+v", song)

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

	if a, ok := song["Name"]; ok && a != "" {
		res = append(res, a)
		hasATN = true
	}
	if a, ok := song["Artist"]; ok && a != "" {
		res = append(res, a)
		hasATN = true
	}
	if a, ok := song["Title"]; ok && a != "" {
		res = append(res, a)
		hasATN = true
	}
	if a, ok := song["Track"]; ok && a != "" {
		res = append(res, a)
	}
	if a, ok := song["Album"]; ok && a != "" {
		res = append(res, a)
	}

	if !hasATN {
		if a, ok := song["File"]; ok && a != "" {
			res = append(res, a)
		}
	}

	s.CurrentSong = strings.Join(res, "; ")
	return
}

func MPDPlay() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		con.Play(-1)
	}
}

func MPDStop() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		con.Stop()
	}
}

func MPDPause() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		if stat, err := con.Status(); err == nil {
			err = con.Pause(stat["state"] != "pause")
		}
	}
}

func MPDNext() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		con.Next()
	}
}

func MPDPrev() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		con.Previous()
	}
}

func MPDVolUp() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		if stat, err := con.Status(); err == nil {
			vol, err := strconv.Atoi(stat["volume"])
			if err != nil {
				return
			}
			vol += 5
			if vol > 100 {
				vol = 100
			}
			con.SetVolume(vol)
		}
	}
}

func MPDVolDown() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		if stat, err := con.Status(); err == nil {
			vol, err := strconv.Atoi(stat["volume"])
			if err != nil {
				return
			}
			vol -= 5
			if vol < 0 {
				vol = 0
			}
			con.SetVolume(vol)
		}
	}
}

func MPDPlaylists() (pls []string) {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		playlists, err := con.ListPlaylists()
		if err == nil {
			for _, pl := range playlists {
				pls = append(pls, pl["playlist"])
			}
		} else {
			log.Printf("MPD.Playlists list error: %s", err)
		}
	} else {
		log.Printf("MPD.Playlists error: %s", err)
	}
	return
}

func MPDPlayPlaylist(playlist string) {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		con.Clear()
		con.PlaylistLoad(playlist, -1, -1)
		con.Play(-1)
	} else {
		log.Printf("MPD.PlayPlaylist error: %s", err)
	}
}

var preMuteVol = -1

func MPDVolMute() {
	con, err := mpd.Dial("tcp", *mpdHost)
	if err == nil {
		defer con.Close()
		if stat, err := con.Status(); err == nil {
			vol, err := strconv.Atoi(stat["volume"])
			if err != nil {
				return
			}
			if vol == 0 {
				if preMuteVol > 0 {
					vol = preMuteVol
				} else {
					vol = 100
				}
			} else {
				preMuteVol = vol
				vol = 0
			}
			con.SetVolume(vol)
		}
	}

}
