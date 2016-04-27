package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"
)

const (
	ActionResultOk = iota
	ActionResultBack
	ActionResultExit
	ActionResultNop
)

const (
	ScreenWidth  = 16
	ScreenHeight = 2
)

type Screen interface {
	Show() []string
	Action(action string) (result int, screen Screen)
}

type TextScreen struct {
	Lines  []string
	offset int
}

func (t *TextScreen) Show() (res []string) {
	for i := t.offset; i < len(t.Lines) && i < (t.offset+ScreenHeight); i++ {
		res = append(res, t.Lines[i])
	}
	for len(res) < ScreenHeight {
		res = append(res, "")
	}
	return
}

func (t *TextScreen) Action(action string) (result int, screen Screen) {
	switch action {
	case configuration.Keys.Menu.Up:
		if t.offset > 0 {
			t.offset--
			return ActionResultOk, nil
		}
	case configuration.Keys.Menu.Down:
		if t.offset+ScreenHeight < len(t.Lines) {
			t.offset++
			return ActionResultOk, nil
		}
	case configuration.Keys.Menu.Back:
		return ActionResultBack, nil
	}
	return
}

type MenuItem struct {
	Label  string
	Cmd    string
	Args   []string
	Kind   string
	Items  []*MenuItem
	offset int
	cursor int
}

func (t *MenuItem) Show() (res []string) {
	for i := t.offset; i < len(t.Items) && i < (t.offset+ScreenHeight); i++ {
		if i == t.cursor {
			res = append(res, "->"+t.Items[i].Label)
		} else {
			res = append(res, "  "+t.Items[i].Label)
		}
	}
	for len(res) < ScreenHeight {
		res = append(res, "")
	}
	return
}

func (t *MenuItem) Action(action string) (result int, screen Screen) {
	switch action {
	case configuration.Keys.Menu.Up:
		t.cursor--
		if t.offset > t.cursor {
			t.offset--
		}
		if t.cursor < 0 {
			t.cursor = len(t.Items) - 1
		}
		if t.offset < 0 {
			t.offset = len(t.Items) - ScreenHeight
			if t.offset < 0 {
				t.offset = 0
			}
		}
		return ActionResultOk, nil
	case configuration.Keys.Menu.Down:
		t.cursor++
		if t.offset < t.cursor-lcdHeight+1 {
			t.offset++
		}
		if t.cursor >= len(t.Items) {
			t.cursor = 0
			t.offset = 0
		}
		return ActionResultOk, nil
	case configuration.Keys.Menu.Select:
		item := t.Items[t.cursor]
		if len(item.Items) > 0 {
			// submenu
			return ActionResultOk, t.Items[t.cursor]
		}
		return item.execute()
	case configuration.Keys.Menu.Back:
		return ActionResultBack, nil
	}
	return
}

func (t *MenuItem) execute() (result int, screen Screen) {
	switch t.Kind {
	case "cmd":
		out, err := exec.Command(t.Cmd, t.Args...).CombinedOutput()
		res := string(out)
		log.Printf("Execute: err=%v, res=%v", err, res)
		lines := strings.Split(res, "\n")
		return ActionResultOk, &TextScreen{Lines: lines}
	}
	return ActionResultNop, nil
}

type StatusScreen struct {
	lastMpdMessage *Status
	last           []string
	Mpd            *MPD
}

func (s *StatusScreen) Show() (res []string) {
	if s.lastMpdMessage == nil || !s.lastMpdMessage.Playing || len(s.last) == 0 {
		log.Printf("lastMpdMessage = %s", s.lastMpdMessage.String())
		n := time.Now()
		res = append(res, loadAvg()+" | stop")
		res = append(res, n.Format("01-02 15:04:05"))
	} else {
		res = s.last[:]
	}
	return
}

func (s *StatusScreen) Action(action string) (result int, screen Screen) {
	switch action {
	case configuration.Keys.MPD.Play:
		s.Mpd.Play()
	case configuration.Keys.MPD.Stop:
		s.Mpd.Stop()
	case configuration.Keys.MPD.Pause:
		s.Mpd.Pause()
	case configuration.Keys.MPD.Next:
		s.Mpd.Next()
	case configuration.Keys.MPD.Prev:
		s.Mpd.Prev()
	case configuration.Keys.MPD.VolUp:
		s.Mpd.VolUp()
	case configuration.Keys.MPD.VolDown:
		s.Mpd.VolDown()
	}
	return ActionResultNop, nil
}

func (s *StatusScreen) MpdUpdate(st *Status) {
	s.lastMpdMessage = st

	if st != nil {
		if st.Error != "" {
			s.last = []string{
				loadAvg() + " | " + st.Status + " " + st.Volume,
				"Err:" + removeNlChars(st.Error),
			}
			return
		}
		if st.Playing {
			if st.Status == "play" {
				s.last = []string{
					loadAvg() + " | play " + st.Flags + " " + st.Volume,
					removeNlChars(st.CurrentSong),
				}
				return
			}
			s.last = []string{
				loadAvg() + " | " + st.Status + " " + st.Volume,
				removeNlChars(st.CurrentSong),
			}
			return
		}
	}

	n := time.Now()
	s.last = []string{
		loadAvg() + " | stop",
		n.Format("01-02 15:04:05"),
	}
}

func loadAvg() string {
	if data, err := ioutil.ReadFile("/proc/loadavg"); err == nil {
		i := bytes.IndexRune(data, ' ')
		if i > 0 {
			return string(data[:i])
		}
	} else {
		log.Printf("main.loadavg error: %v", err)
	}
	return ""
}
