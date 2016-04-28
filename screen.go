package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	ActionResultOk = iota
	ActionResultBack
	ActionResultExit
	ActionResultNop
)

const (
	CharPlay  = "\b00010000"
	CharPause = "\b10100000"
	CharStop  = "\b10110111"
	CharError = "!"
)

// Screen define single screen for display
type Screen interface {
	// Show return lines to display
	Show() []string
	// Action perform some action on screen
	Action(action string) (result int, screen Screen)
}

type TextScreen struct {
	Lines  []string
	offset int
}

func (t *TextScreen) Show() (res []string) {
	if len(t.Lines) == 0 {
		res = append(res, "No text")
	} else {
		for i := t.offset; i < len(t.Lines) && i < (t.offset+lcdHeight); i++ {
			res = append(res, t.Lines[i])
		}
	}
	for len(res) < lcdHeight {
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
		if t.offset+lcdHeight < len(t.Lines) {
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
	for i := t.offset; i < len(t.Items) && i < (t.offset+lcdHeight); i++ {
		if i == t.cursor {
			res = append(res, "->"+t.Items[i].Label)
		} else {
			res = append(res, "  "+t.Items[i].Label)
		}
	}
	for len(res) < lcdHeight {
		res = append(res, "")
	}
	return
}

func (t *MenuItem) Action(action string) (result int, screen Screen) {
	switch action {
	case configuration.Keys.Menu.Up:
		t.cursor, t.offset = cursorScrollUp(
			t.cursor, t.offset, len(t.Items))
		return ActionResultOk, nil
	case configuration.Keys.Menu.Down:
		t.cursor, t.offset = cursorScrollDown(
			t.cursor, t.offset, len(t.Items))
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
		res := strings.TrimSpace(string(out))
		if res == "" {
			res = "<no output>"
		}
		log.Printf("Execute: err=%v, res=%v", err, res)
		lines := strings.Split(res, "\n")
		return ActionResultOk, &TextScreen{Lines: lines}
	case "mpd":
		switch t.Cmd {
		case "playlists":
			return ActionResultOk, NewMPDPlaylistScreen()
		}
	}
	return ActionResultNop, nil
}

type StatusScreen struct {
	lastMpdMessage *Status
	last           []string
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
		MPDPlay()
	case configuration.Keys.MPD.Stop:
		MPDStop()
	case configuration.Keys.MPD.Pause:
		MPDPause()
	case configuration.Keys.MPD.Next:
		MPDNext()
	case configuration.Keys.MPD.Prev:
		MPDPrev()
	case configuration.Keys.MPD.VolUp:
		MPDVolUp()
	case configuration.Keys.MPD.VolDown:
		MPDVolDown()
	}
	return ActionResultNop, nil
}

func (s *StatusScreen) MpdUpdate(st *Status) {
	s.lastMpdMessage = st

	if st != nil {
		if st.Error != "" {
			s.last = []string{
				loadAvg() + " | " + mpdStatusToStr(st.Status) + " " + st.Volume,
				"Err:" + removeNlChars(st.Error),
			}
			return
		}
		if st.Status != "stop" {
			s.last = []string{
				loadAvg() + " | " + mpdStatusToStr(st.Status) + st.Flags + " " + st.Volume,
				removeNlChars(st.CurrentSong),
			}
			return
		}
	}

	n := time.Now()
	s.last = []string{
		loadAvg() + " | " + mpdStatusToStr(st.Status),
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

type UrgentMsgScreen struct {
	mu       sync.Mutex
	messages [][]string
	offset   int
}

func (u *UrgentMsgScreen) HasMessages() bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	return len(u.messages) > 0
}

func (u *UrgentMsgScreen) AddMsg(msg []string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.messages = append(u.messages, msg)
	log.Printf("AddMsg: %#v", u.messages)
}

func (u *UrgentMsgScreen) Show() (res []string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.messages) == 0 {
		res = append(res, "No messages")
	} else {
		msg := u.messages[0]
		for i := u.offset; i < len(msg) && i < (u.offset+lcdHeight); i++ {
			res = append(res, msg[i])
		}
	}
	for len(res) < lcdHeight {
		res = append(res, "")
	}
	return
}

func (u *UrgentMsgScreen) Action(action string) (result int, screen Screen) {
	u.mu.Lock()
	defer u.mu.Unlock()

	switch action {
	case configuration.Keys.Menu.Up:
		if len(u.messages) > 0 && u.offset > 0 {
			u.offset--
			return ActionResultOk, nil
		}
	case configuration.Keys.Menu.Down:
		if len(u.messages) > 0 && u.offset+lcdHeight < len(u.messages[0]) {
			u.offset++
			return ActionResultOk, nil
		}
	case configuration.Keys.Menu.Select, configuration.Keys.Menu.Back:
		u.offset = 0
		if len(u.messages) > 1 {
			u.messages = u.messages[1:]
			return ActionResultOk, nil
		}
		u.messages = nil
		return ActionResultBack, nil
	}
	return ActionResultNop, nil
}

type MPDPlaylistsScreen struct {
	offset    int
	cursor    int
	playlists []string
}

func NewMPDPlaylistScreen() *MPDPlaylistsScreen {
	return &MPDPlaylistsScreen{
		playlists: MPDPlaylists(),
	}
}

func (m *MPDPlaylistsScreen) Show() (res []string) {
	if len(m.playlists) == 0 {
		res = append(res, "No playlists")
	} else {
		for i := m.offset; i < len(m.playlists) && i < (m.offset+lcdHeight); i++ {
			if i == m.cursor {
				res = append(res, "->"+m.playlists[i])
			} else {
				res = append(res, "  "+m.playlists[i])
			}
		}
	}
	for len(res) < lcdHeight {
		res = append(res, "")
	}
	return
}

func (m *MPDPlaylistsScreen) Action(action string) (result int, screen Screen) {
	switch action {
	case configuration.Keys.Menu.Up:
		m.cursor, m.offset = cursorScrollUp(m.cursor, m.offset, len(m.playlists))
		return ActionResultOk, nil
	case configuration.Keys.Menu.Down:
		m.cursor, m.offset = cursorScrollDown(m.cursor, m.offset, len(m.playlists))
		return ActionResultOk, nil
	case configuration.Keys.Menu.Select:
		playlist := m.playlists[m.cursor]
		MPDPlayPlaylist(playlist)
		return ActionResultOk, nil
	case configuration.Keys.Menu.Back:
		return ActionResultBack, nil
	}
	return
}

func cursorScrollUp(cursor, offset, items int) (rcursor, roffset int) {
	cursor--
	if offset > cursor {
		offset--
	}
	if cursor < 0 {
		cursor = items - 1
	}
	if offset < 0 {
		offset = items - lcdHeight
		if offset < 0 {
			offset = 0
		}
	}
	return cursor, offset
}

func cursorScrollDown(cursor, offset, items int) (rcursor, roffset int) {
	cursor++
	if offset < cursor-lcdHeight+1 {
		offset++
	}

	if cursor >= items {
		cursor = 0
		offset = 0
	}
	return cursor, offset
}

func mpdStatusToStr(status string) string {
	switch status {
	case "play":
		return CharPlay
	case "pause":
		return CharPause
	}
	return CharStop
}
