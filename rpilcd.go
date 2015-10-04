package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var AppVersion = "dev"

func main() {

	flag.Parse()

	lcd := InitLcd()
	lcd.LcdString("\n")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for _ = range c {
			//lcd.Close()
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
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case msg := <-mpd.Message:
			ts.Set(msg.String())
			//log.Printf("main.msg: %v", msg)
		case <-ticker.C:
			text := ts.Tick()
			log.Printf("main.msg: %v", text)
			lcd.LcdString(text)
		}
	}
}

type textScrollerLine struct {
	lineOrg    string
	line       string
	needScroll bool
}

func (tsl *textScrollerLine) set(inp string, width int) {
	log.Printf("textScrollerLine.set : %+v", inp)
	if inp == tsl.lineOrg {
		return
	}
	tsl.lineOrg = inp
	tsl.needScroll = len(inp) > width
	if tsl.needScroll {
		tsl.line = inp + " | "
	} else {
		if len(inp) < width {
			inp += strings.Repeat(" ", width-len(inp))
		}
		tsl.line = inp
	}
}

func (tsl *textScrollerLine) scroll() string {
	if tsl.needScroll {
		line := tsl.line[1:] + string(tsl.line[0])
		tsl.line = line
	}
	return tsl.line
}

type TextScroller struct {
	Width int
	lines []*textScrollerLine

	mu sync.Mutex
}

func NewTextScroller(width int) *TextScroller {
	res := &TextScroller{
		Width: width,
	}
	for i := 0; i < 2; i++ {
		l := &textScrollerLine{}
		l.set("  ", width)
		res.lines = append(res.lines, l)
	}
	return res
}

func (t *TextScroller) Set(text string) {
	log.Printf("TextScroller.Set: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	for line, l := range strings.Split(text, "\n") {
		if line >= 2 {
			return
		}
		t.lines[line].set(l, t.Width)
	}
}

func (t *TextScroller) Tick() (res string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, l := range t.lines {
		res += l.scroll()[:t.Width] + "\n"
	}
	return strings.TrimRight(res, "\n")
}
