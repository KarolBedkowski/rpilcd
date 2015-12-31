package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	//_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const lcdWidth = 16

// AppVersion global var
var AppVersion = "dev"

// Display (output) interface
type Display interface {
	Display(string)
	Close()
}

func main() {

	//go func() {
	//log.Println(http.ListenAndServe(":6060", nil))
	//}()

	log.Printf("RPI LCD ver %s starting...", AppVersion)

	soutput := flag.Bool("console", false, "Print on console instead of lcd")
	refreshInt := flag.Int64("interval", 1000, "Interval between lcd updates in ms")
	startWS := flag.Bool("start_ws", true, "Start WS for external content")
	wsAddr := flag.String("ws_addr", "localhost:8681", "Webservice address")

	flag.Parse()

```	ws := WSServer{
		Addr: *wsAddr,
	}
	if *startWS {
		ws.Start()
	}

	log.Printf("main: interval: %d ms", *refreshInt)

	var disp Display
	if *soutput {
		log.Printf("main: starting console")
		disp = NewConsole()
	} else {
		log.Printf("main: starting lcd")
		disp = NewLcd()
	}

	disp.Display(" \n ")
	mpd := NewMPD()

	defer func() {
		if e := recover(); e != nil {
			log.Printf("Recover: %v", e)
		}
		log.Printf("main.defer: closing disp")
		disp.Close()
		log.Printf("main.defer: closing mpd")
		mpd.Close()
		time.Sleep(2 * time.Second)
		log.Printf("main.defer: all closed")
	}()

	mpd.Connect()

	time.Sleep(1 * time.Second)

	log.Printf("main: entering loop")

	ts := NewTextScroller(lcdWidth)
	pt := NewPrioText(lcdWidth)
	pt.PrioMsgTime = 5000 / int(*refreshInt)
	ticker := time.NewTicker(time.Duration(*refreshInt) * time.Millisecond)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	lastMpdMessage := mpd.GetStatus()
	for {
		select {
		case _ = <-sig:
			return
		case msg := <-ws.Message:
			pt.Add(msg)
		case msg := <-mpd.Message:
			lastMpdMessage = msg
			ts.Set(formatData(lastMpdMessage))
		case <-ticker.C:
			text, ok := pt.Get()
			if !ok {
				if lastMpdMessage == nil || !lastMpdMessage.Playing {
					log.Printf("lastMpdMessage = %s", lastMpdMessage.String())
					n := time.Now()
					ts.Set(loadAvg() + " | stop\n " + n.Format("01-02 15:04:05"))
				}
				text = ts.Tick()
			}
			disp.Display(text)
		}
	}
}

func formatData(s *Status) string {
	if s != nil && s.Playing {
		if s.Status == "play" {
			return loadAvg() + " | " + s.Flags + s.Volume + "\n" + removeNlChars(s.CurrentSong)
		}
		return loadAvg() + " | " + s.Status + " " + s.Volume + "\n" + removeNlChars(s.CurrentSong)
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
