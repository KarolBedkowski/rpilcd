package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

const minPressInterval = time.Duration(500) * time.Millisecond

type DisplayP struct {
	lastMpdMessage *Status
	um             *UrgentMsgManager
	ts             *TextScroller
	disp           Display
	menu           []*Menu
	menuCursor     []int
	menuOffset     []int
	lastPressTime  time.Time
}

func NewDisplayP(console bool, refreshInt int) *DisplayP {
	d := &DisplayP{}
	if console {
		log.Printf("main: starting console")
		d.disp = NewConsole()
	} else {
		log.Printf("main: starting lcd")
		d.disp = NewLcd()
	}

	d.disp.Display(" \n ")

	d.ts = NewTextScroller(lcdWidth)
	d.um = NewUrgentMsgManager(lcdWidth)
	d.um.DefaultTimeout = 5000 / int(refreshInt)

	return d
}

func (d *DisplayP) MenuDisplayed() bool {
	return d.menu != nil && len(d.menu) > 0
}

func (d *DisplayP) Close() {
	d.disp.Close()
}

func (d *DisplayP) NewMsg(msg *Status) {
	d.lastMpdMessage = msg
	d.ts.Set(formatData(msg))
}

func (d *DisplayP) NewCommand(msg string) {
	now := time.Now()
	if now.Sub(d.lastPressTime) < minPressInterval {
		return
	}
	d.lastPressTime = now

	msg = strings.TrimSpace(msg)
	log.Printf("NewCommand '%s'", msg)
	switch msg {
	case configuration.Keys.Menu.Up:
		d.MenuUp()
		return
	case configuration.Keys.Menu.Down:
		d.MenuDown()
		return
	case configuration.Keys.Menu.Show:
		d.ShowMenu()
		return
	case configuration.Keys.Menu.Select:
		d.MenuSelect()
		return
	case configuration.Keys.Menu.Back:
		d.MenuBack()
		return
	}
	if strings.HasPrefix(msg, "toggle-backlight") {
		d.disp.ToggleBacklight()
	} else {
		d.um.AddJSON(msg)
	}
}

func (d *DisplayP) DisplayStatus() {
	// do not show status when menu displayed
	if d.MenuDisplayed() {
		return
	}
	text, ok := d.um.Get()
	if !ok {
		if d.lastMpdMessage == nil || !d.lastMpdMessage.Playing {
			log.Printf("lastMpdMessage = %s", d.lastMpdMessage.String())
			n := time.Now()
			d.ts.Set(loadAvg() + " | stop\n " + n.Format("01-02 15:04:05"))
		}
		text = d.ts.Tick()
	}
	d.disp.Display(text)
}

func (d *DisplayP) ShowMenu() {
	if d.MenuDisplayed() {
		return
	}
	d.menu = append(d.menu, &configuration.Menu)
	d.menuCursor = append(d.menuCursor, 0)
	d.menuOffset = append(d.menuOffset, 0)
	d.DisplayMenu()
}

func (d *DisplayP) MenuUp() {
	menu := d.menu[len(d.menu)-1]
	cursor := d.menuCursor[len(d.menuCursor)-1]
	offset := d.menuOffset[len(d.menuOffset)-1]
	cursor--
	if cursor < 0 {
		cursor = len(menu.Items) - 1
		offset = cursor - lcdHeight + 1
		if offset < 0 {
			offset = 0
		}
	} else {
		if cursor < offset {
			offset--
		}
	}
	d.menuCursor[len(d.menuCursor)-1] = cursor
	d.menuOffset[len(d.menuOffset)-1] = offset
	d.DisplayMenu()
}

func (d *DisplayP) MenuDown() {
	menu := d.menu[len(d.menu)-1]
	cursor := d.menuCursor[len(d.menuCursor)-1]
	offset := d.menuOffset[len(d.menuOffset)-1]
	cursor++
	if cursor >= len(menu.Items) {
		cursor = 0
		offset = 0
	} else {
		if cursor >= offset+lcdHeight {
			offset++
		}
	}
	d.menuCursor[len(d.menuCursor)-1] = cursor
	d.menuOffset[len(d.menuOffset)-1] = offset
	d.DisplayMenu()
}

func (d *DisplayP) MenuBack() {
	if !d.MenuDisplayed() {
		return
	}
	cnt := len(d.menu)
	if cnt == 1 {
		d.menu = nil
		d.menuOffset = nil
		d.menuCursor = nil
		d.DisplayStatus()
		return
	}
	d.menu = d.menu[:cnt-1]
	d.menuOffset = d.menuOffset[:cnt-1]
	d.menuCursor = d.menuCursor[:cnt-1]
	d.DisplayMenu()
}

func (d *DisplayP) MenuSelect() {
	if !d.MenuDisplayed() {
		return
	}
	menu := d.menu[len(d.menu)-1]
	cursor := d.menuCursor[len(d.menuCursor)-1]
	subMenu := menu.Items[cursor]
	if subMenu.Items != nil && len(subMenu.Items) > 0 {
		// enter in submenu
		d.menu = append(d.menu, subMenu)
		d.menuCursor = append(d.menuCursor, 0)
		d.menuOffset = append(d.menuOffset, 0)
		d.DisplayMenu()
	} else {
		log.Printf("execute: %#v", subMenu)
		task := Task{
			Kind: subMenu.Kind,
			Cmd:  subMenu.Cmd,
			Args: subMenu.Args,
		}
		res, _ := task.Execute()
		d.disp.Display(res)
	}
}

func (d *DisplayP) DisplayMenu() {
	if !d.MenuDisplayed() {
		d.disp.Display("no menu")
		return
	}
	menu := d.menu[len(d.menu)-1]

	if menu.Items != nil && len(menu.Items) > 0 {
		// menu has items; show
		cursor := d.menuCursor[len(d.menuCursor)-1]
		offset := d.menuOffset[len(d.menuOffset)-1]

		rows := menu.GetItems(lcdHeight, offset, cursor)
		text := strings.Join(rows, "\n")
		d.disp.Display(text)
	}
}

func formatData(s *Status) string {
	if s != nil {
		if s.Error != "" {
			return loadAvg() + " | " + s.Status + " " + s.Volume + "\nErr:" + removeNlChars(s.Error)
		}
		if s.Playing {
			if s.Status == "play" {
				return loadAvg() + " | play " + s.Flags + " " + s.Volume + "\n" + removeNlChars(s.CurrentSong)
			}
			return loadAvg() + " | " + s.Status + " " + s.Volume + "\n" + removeNlChars(s.CurrentSong)
		}
	}

	n := time.Now()
	return loadAvg() + " | stop\n " + n.Format("01-02 15:04:05")
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
