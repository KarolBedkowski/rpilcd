package main

import (
	"log"
	"strings"
	"time"
)

const minPressInterval = time.Duration(500) * time.Millisecond

type DisplayP struct {
	um            *UrgentMsgManager
	ts            *TextScroller
	disp          Display
	statusScr     *StatusScreen
	screens       []Screen
	lastPressTime time.Time
}

func NewDisplayP(console bool, refreshInt int, mpd *MPD) *DisplayP {
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
	d.statusScr = &StatusScreen{Mpd: mpd}

	return d
}

func (d *DisplayP) MenuDisplayed() bool {
	return len(d.screens) > 0
}

func (d *DisplayP) Close() {
	d.disp.Close()
}

func (d *DisplayP) NewCommand(msg string) {
	now := time.Now()
	if now.Sub(d.lastPressTime) < minPressInterval {
		return
	}
	d.lastPressTime = now

	msg = strings.TrimSpace(msg)
	log.Printf("NewCommand '%s' '%s'", msg,
		configuration.Keys.Menu.Show)

	if msg == configuration.Keys.Menu.Show {
		if len(d.screens) > 0 {
			d.screens = nil
			d.display()
			return
		}
		d.screens = append(d.screens, configuration.Menu)
		d.display()
		return
	}

	screen := d.currentScreen()
	res, nextScreen := screen.Action(msg)
	switch res {
	case ActionResultBack:
		d.screens = d.screens[:len(d.screens)-1]
		d.display()
		return
	case ActionResultExit:
		d.screens = nil
		d.display()
		return
	case ActionResultOk:
		if nextScreen != nil {
			d.screens = append(d.screens, nextScreen)
			d.display()
			return
		}
	}

	d.um.AddJSON(msg)
}

func (d *DisplayP) currentScreen() Screen {
	if len(d.screens) > 0 {
		return d.screens[len(d.screens)-1]
	}
	return d.statusScr

}

func (d *DisplayP) display() {
	screen := d.currentScreen()
	lines := screen.Show()
	text := strings.Join(lines, "\n")
	d.ts.Set(text)
	d.disp.Display(d.ts.Tick())
}

func (d *DisplayP) UpdateMpdStatus(status *Status) {
	d.statusScr.MpdUpdate(status)
}

func (d *DisplayP) Tick() {
	d.display()
}
