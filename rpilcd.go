package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	//"net/http"
	//_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var AppVersion = "dev"

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

	flag.Parse()

	log.Printf("Interval: %d ms", *refreshInt)

	var disp Display
	if *soutput {
		disp = NewConsole()
	} else {
		disp = NewLcd()
	}

	disp.Display(" \n ")
	mpd := NewMPD()

	defer func() {
		if e := recover(); e != nil {
			log.Printf("Recover: %v", e)
		}
		log.Printf("Closing disp")
		disp.Close()
		log.Printf("Closing mpd")
		mpd.Close()
		time.Sleep(2 * time.Second)
		log.Printf("Closed")
	}()

	mpd.Connect()

	time.Sleep(1 * time.Second)

	log.Printf("main: entering loop")

	ts := NewTextScroller(LCD_WIDTH)
	ticker := time.NewTicker(time.Duration(*refreshInt) * time.Millisecond)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	//	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSTOP,
	//		syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGSEGV, syscall.SIGCHLD)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case msg := <-mpd.Message:
			ts.Set(formatData(msg))
		case <-ticker.C:
			text := ts.Tick()
			disp.Display(text)
		case sig := <-sig:
			log.Printf("Signal %v", sig)
			return
		}
	}
}

func formatData(s *Status) string {
	if s.Playing {
		if s.Status == "play" {
			return loadAvg() + " | " + s.Flags + s.Volume + "\n" + removeNlChars(s.CurrentSong)
		} else {
			return loadAvg() + " | " + s.Status + " " + s.Volume + "\n" + removeNlChars(s.CurrentSong)
		}
	}

	n := time.Now()
	return loadAvg() + " | " + s.Status + "\n " + n.Format("01-02 15:04:05")
}

func loadAvg() string {
	if data, err := ioutil.ReadFile("/proc/loadavg"); err == nil {
		i := bytes.IndexRune(data, ' ')
		if i > 0 {
			return string(data[:i])
		}
	} else {
		log.Printf("loadavg errorL %v", err)
	}
	return ""
}
