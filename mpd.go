package main

import (
	"flag"
	"github.com/turbowookie/gompd/mpd"
	"log"
	"strconv"
	"strings"
)

var (
	mpdHost = flag.String("mpdHost", "pi:6600", "MPD address")
)

type Status struct {
	CurrentSong string
	Playing     bool
	Status      string
	Flags       string
	Volume      string
}

func (s *Status) String() string {
	if s.Playing {
		if s.Status == "play" {
			return s.Flags + s.Volume + "\n" + s.CurrentSong
		} else {
			return s.Status + " " + s.Volume + "\n" + s.CurrentSong
		}
	}
	return s.Status
}

type MPD struct {
	Message    chan *Status
	connection *mpd.Client
	watcher    *mpd.Watcher
}

func NewMPD() *MPD {
	return &MPD{
		Message: make(chan *Status, 10),
	}
}

func (m *MPD) watch() (err error) {
	m.watcher, err = mpd.NewWatcher("tcp", *mpdHost, "")
	if err != nil {
		log.Printf("mpd.watcher: connect to %v error: %v", *mpdHost, err.Error())
		return err
	}
	log.Printf("mpd.watcher: connected to %v", *mpdHost)

	defer func() {
		log.Println("mpd.watcher: closing")
		m.watcher.Close()
		m.watcher = nil
	}()

	log.Println("mpd.watcher: starting watch")

	m.Message <- m.getStatus()

	for {
		select {
		case subsystem := <-m.watcher.Event:
			log.Printf("mpd.watcher - event: %v", subsystem)
			switch subsystem {
			case "player":
				m.Message <- m.getStatus()
			}
		case err := <-m.watcher.Error:
			log.Printf("mpd.watcher: error event: %v", err)
			return nil
		}
	}

	return nil
}

func (m *MPD) Connect() (err error) {
	if m.connection != nil {
		if err = m.connection.Ping(); err != nil {
			m.connection.Close()
			m.connection = nil
		}
	}
	if m.connection == nil {
		m.connection, err = mpd.Dial("tcp", *mpdHost)
		if err != nil {
			log.Printf("mpd.Connect: connect do %s error: %v", *mpdHost, err.Error())
			return err
		}
		log.Printf("mpd.Connect: connected to %s", *mpdHost)
	}
	err = m.watch()
	if err != nil {
		log.Printf("mpd.Connect: start watch error: %v", err)
	}
	return
}

func (m *MPD) getStatus() (s *Status) {
	s = &Status{
		Playing: false,
		Flags:   "ERR",
	}

	song, err := m.connection.CurrentSong()
	if err != nil {
		log.Printf("mpd.getStatus: CurrentSong error: %v", err.Error())
		return
	}

	status, err := m.connection.Status()
	if err != nil {
		log.Printf("mpd.getStatus: Status error: %v", err.Error())
		return
	}

	s.Status = status["state"]
	s.Playing = s.Status != "stop"
	s.Volume = status["volume"]
	s.Flags = ""

	if status["repeat"] != "0" {
		s.Flags += "R"
	}
	if status["random"] != "0" {
		s.Flags += "S"
	}

	//log.Printf("Status: %+v", status)
	//log.Printf("Song: %+v", song)

	res := make([]string, 0)

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

	s.CurrentSong = strings.Join(res, " - ")
	return
}
