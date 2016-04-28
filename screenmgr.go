package main

import (
	"log"
	"strings"
	"time"
)

const minCmdsInterval = time.Duration(500) * time.Millisecond

type ScreenMgr struct {
	ums         *UrgentMsgScreen
	ts          *TextScroller
	disp        Display
	statusScr   *StatusScreen
	screens     []Screen
	lastCmdTime time.Time
}

func NewScreenMgr(console bool, refreshInt int) *ScreenMgr {
	d := &ScreenMgr{}
	if console {
		log.Printf("main: starting console")
		d.disp = NewConsole()
	} else {
		log.Printf("main: starting lcd")
		d.disp = NewLcd()
	}

	d.disp.Display(" \n ")

	d.ts = NewTextScroller(lcdWidth, lcdHeight)
	d.statusScr = &StatusScreen{}
	d.ums = &UrgentMsgScreen{}

	return d
}

func (d *ScreenMgr) Close() {
	d.disp.Close()
}

func (d *ScreenMgr) NewCommand(msg string) {
	now := time.Now()
	if now.Sub(d.lastCmdTime) < minCmdsInterval {
		return
	}
	d.lastCmdTime = now

	msg = strings.TrimSpace(msg)
	log.Printf("NewCommand '%s'", msg)

	// globla commands
	switch msg {
	// toggle menu
	case configuration.Keys.Menu.Show:
		if len(d.screens) > 0 {
			d.screens = nil
			d.display()
			return
		}
		d.screens = append(d.screens, configuration.Menu)
		d.display()
		return
	case configuration.Keys.ToggleLCD:
		d.disp.ToggleBacklight()
		return
	}

	screen := d.currentScreen()
	log.Printf("current screen: %#v", screen)
	res, nextScreen := screen.Action(msg)
	switch res {
	case ActionResultBack:
		if len(d.screens) > 0 {
			d.screens = d.screens[:len(d.screens)-1]
		}
		d.display()
		return
	case ActionResultExit:
		d.screens = nil
		d.display()
		return
	case ActionResultOk:
		if nextScreen != nil {
			d.screens = append(d.screens, nextScreen)
		}
		d.display()
		return
	}
	d.AddUrgentMsg(msg)
}

func (d *ScreenMgr) currentScreen() Screen {
	if d.ums.HasMessages() {
		return d.ums
	}
	if len(d.screens) > 0 {
		return d.screens[len(d.screens)-1]
	}
	return d.statusScr

}

func (d *ScreenMgr) display() {
	screen := d.currentScreen()
	lines := screen.Show()
	text := strings.Join(lines, "\n")
	d.ts.Set(text)
	d.disp.Display(d.ts.Tick())
}

func (d *ScreenMgr) UpdateMpdStatus(status *Status) {
	d.statusScr.MpdUpdate(status)
}

func (d *ScreenMgr) AddUrgentMsg(msg string) {
	d.ums.AddMsg(strings.Split(msg, "\n"))
}

func (d *ScreenMgr) Tick() {
	d.display()
}
