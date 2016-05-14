package main

import (
	"github.com/golang/glog"
	"net/http"
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
	lastContent string
}

func NewScreenMgr(console bool) *ScreenMgr {
	d := &ScreenMgr{}
	if console {
		glog.Infof("main: starting console")
		d.disp = NewConsole()
	} else {
		glog.Infof("main: starting lcd")
		if lcd := NewLcd(); lcd != nil {
			d.disp = lcd
		} else {
			glog.Infof("main: fail back to console")
			d.disp = NewConsole()
		}
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
	glog.Infof("NewCommand '%s'", msg)

	// globla commands
	switch msg {
	// toggle menu
	case configuration.Keys.Menu.Show:
		if len(d.screens) > 0 {
			d.screens = nil
		} else {
			d.screens = append(d.screens, configuration.Menu)
		}
		d.display(false)
		return
	case configuration.Keys.ToggleLCD:
		d.disp.ToggleBacklight()
		return
	}

	screen := d.currentScreen()
	if glog.V(1) {
		glog.Infof("current screen: %#v", screen)
	}
	res, nextScreen := screen.Action(msg)
	switch res {
	case ActionResultBack:
		if len(d.screens) > 0 {
			d.screens = d.screens[:len(d.screens)-1]
		}
		d.display(false)
	case ActionResultExit:
		d.screens = nil
		d.display(false)
	case ActionResultOk:
		if nextScreen != nil {
			d.screens = append(d.screens, nextScreen)
		}
		d.display(false)
	default:
		d.AddUrgentMsg(msg)
	}
	d.lastCmdTime = time.Now()
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

func (d *ScreenMgr) display(tick bool) {
	screen := d.currentScreen()
	if !screen.Valid() {
		if len(d.screens) > 0 {
			d.screens = d.screens[:len(d.screens)-1]
		}
		screen = d.currentScreen()
	}
	lines, fixPart := screen.Show()
	text := strings.Join(lines, "\n")
	d.lastContent = text
	d.ts.Set(text, fixPart)
	if tick {
		d.disp.Display(d.ts.Tick())
	} else {
		d.disp.Display(d.ts.Get())
	}
}

func (d *ScreenMgr) UpdateMpdStatus(status *MPDStatus) {
	d.statusScr.MpdUpdate(status)
}

func (d *ScreenMgr) AddUrgentMsg(msg string) {
	d.ums.AddMsg(strings.Split(msg, "\n"))
}

func (d *ScreenMgr) Tick() {
	d.display(true)
}

func (d *ScreenMgr) WebHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(d.lastContent))
}
