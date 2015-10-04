package main

import (
	"log"
	"strings"
	"sync"
)

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
		l.set(strings.Repeat(" ", width), width)
		res.lines = append(res.lines, l)
	}
	return res
}

func (t *TextScroller) Set(text string) {
	log.Printf("TextScroller.Set: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	for line, l := range strings.SplitN(text, "\n", 2) {
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
