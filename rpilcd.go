package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	//_ "net/http/pprof"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const lcdWidth = 16
const lcdHeight = 2

// AppVersion global var
var AppVersion = "dev"

// Display (output) interface
type Display interface {
	Display(string)
	Close()
	ToggleBacklight()
}

func main() {

	//go func() {
	//log.Println(http.ListenAndServe(":6060", nil))
	//}()

	log.Printf("RPI LCD ver %s starting...", AppVersion)

	soutput := flag.Bool("console", false, "Print on console instead of lcd")
	refreshInt := flag.Int64("interval", 1000, "Interval between lcd updates in ms")
	satartService := flag.Bool("startService", true, "Start TCP server for urgent messages")
	serviceAddr := flag.String("serviceAddr", "localhost:8681", "TCP server address")
	httpServAddr := flag.String("listen-address", ":8001", "The address to listen on for HTTP requests.")

	flag.Parse()

	ws := UMServer{
		Addr: *serviceAddr,
	}
	if *satartService {
		ws.Start()
	}

	if *httpServAddr != "" {
		http.Handle("/metrics", prometheus.Handler())
		go http.ListenAndServe(*httpServAddr, nil)
	}

	err := loadMenu("menu.toml")
	log.Printf("Load menu: %s", err)
	log.Printf("menu: %s", mainMenu)

	log.Printf("main: interval: %d ms", *refreshInt)

	mpd := NewMPD()
	dispP := NewDisplayP(*soutput, int(*refreshInt))

	lirc := NewLirc()

	defer func() {
		//if e := recover(); e != nil {
		//	log.Printf("Recover: %v", e)
		//}
		log.Printf("main.defer: closing disp")
		dispP.Close()
		log.Printf("main.defer: closing mpd")
		mpd.Close()
		lirc.Close()
		time.Sleep(2 * time.Second)
		log.Printf("main.defer: all closed")
	}()

	mpd.Connect()
	dispP.lastMpdMessage = mpd.GetStatus()

	time.Sleep(1 * time.Second)

	log.Printf("main: entering loop")

	ticker := time.NewTicker(time.Duration(*refreshInt) * time.Millisecond)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		// TODO: handle lirc event
		select {
		case _ = <-sig:
			return
		case ev := <-lirc.Events:
			{
				log.Printf("lirc event: %s", ev)
				switch ev {
				case "KEY_4":
					dispP.ShowMenu()
				case "KEY_5":
					dispP.MenuBack()
				}
			}
		case msg := <-ws.Message:
			dispP.NewCommand(msg)
		case msg := <-mpd.Message:
			dispP.NewMsg(msg)
		case <-ticker.C:
			dispP.DisplayStatus()
		}
	}
}

type DisplayP struct {
	lastMpdMessage *Status
	um             *UrgentMsgManager
	ts             *TextScroller
	disp           Display
	menu           []*Menu
	menuCursor     []int
	menuOffset     []int
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

func (d *DisplayP) Close() {
	d.disp.Close()
}

func (d *DisplayP) NewMsg(msg *Status) {
	d.lastMpdMessage = msg
	d.ts.Set(formatData(msg))
}

func (d *DisplayP) NewCommand(msg string) {
	msg = strings.TrimSpace(msg)
	log.Printf("NewCommand '%s'", msg)
	if msg == "mup" {
		d.MenuUp()
		return
	}
	if msg == "mdown" {
		d.MenuDown()
		return
	}
	if msg == "menu" {
		d.ShowMenu()
		return
	}
	if msg == "msel" {
		d.MenuSelect()
		return
	}
	if msg == "mback" {
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
	if d.menu != nil && len(d.menu) > 0 {
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
	if d.menu != nil && len(d.menu) > 0 {
		return
	}
	d.menu = append(d.menu, &mainMenu.Menu)
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
		if cursor > offset+lcdHeight {
			offset++
		}
	}
	d.menuCursor[len(d.menuCursor)-1] = cursor
	d.menuOffset[len(d.menuOffset)-1] = offset
	d.DisplayMenu()
}

func (d *DisplayP) MenuBack() {
	if d.menu == nil || len(d.menu) == 0 {
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
	if d.menu == nil || len(d.menu) == 0 {
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
		// execute action
	}
}

func (d *DisplayP) DisplayMenu() {
	if d.menu == nil && len(d.menu) == 0 {
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
