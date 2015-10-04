package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var AppVersion = "dev"

type Display interface {
	Display(string)
	Close()
}

type SimpleDisplay struct{}

func (sd *SimpleDisplay) Display(text string) {
	for i, l := range strings.Split(text, "\n") {
		log.Printf("SimpleDisplay: [%d] '%s'", i, l)
	}
}

func (sd *SimpleDisplay) Close() {
}

func main() {
	log.Printf("RPI LCD ver %s starting...", AppVersion)

	soutput := flag.Bool("console", false, "Print on console instead of lcd")
	refreshInt := flag.Int64("interval", 500, "Interval between lcd updates in ms")

	flag.Parse()

	log.Printf("Interval: %d ms", *refreshInt)

	var disp Display
	if *soutput {
		disp = &SimpleDisplay{}
	} else {
		disp = InitLcd()
	}

	disp.Display("\n")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for _ = range c {
			disp.Close()
			os.Exit(0)
		}
	}()

	mpd := NewMPD()
	go func() {
		for {
			if mpd.Connect() == nil {
				log.Printf("main: mpd connected")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	time.Sleep(1 * time.Second)

	log.Printf("main: entering loop")

	ts := NewTextScroller(LCD_WIDTH)
	ticker := time.NewTicker(time.Duration(*refreshInt) * time.Millisecond)

	for {
		select {
		case msg := <-mpd.Message:
			ts.Set(formatData(msg))
		case <-ticker.C:
			text := ts.Tick()
			disp.Display(text)
		}
	}
}

func formatData(s *Status) string {
	if s.Playing {
		if s.Status == "play" {
			return loadAvg() + " | " + s.Flags + s.Volume + "\n" + s.CurrentSong
		} else {
			return loadAvg() + " | " + s.Status + " " + s.Volume + "\n" + s.CurrentSong
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
