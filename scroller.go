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

// TextScroller format some text to display in few character display
type TextScroller struct {
	Width         int
	lines         []*textScrollerLine
	prioLines     []string
	prioLinesTick int
	PrioMsgTime   int

	mu sync.Mutex
}

// NewTextScroller create new TextScroller struct
func NewTextScroller(width int) *TextScroller {
	res := &TextScroller{
		Width:     width,
		prioLines: make([]string, 0),
	}
	for i := 0; i < 2; i++ {
		l := &textScrollerLine{}
		l.set(strings.Repeat(" ", width), width)
		res.lines = append(res.lines, l)
	}
	return res
}

// Set put some text to TextScroller
func (t *TextScroller) Set(text string) {
	log.Printf("TextScroller.Set: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	for line, l := range strings.SplitN(text, "\n", 2) {
		t.lines[line].set(l, t.Width)
	}
}

func (t *TextScroller) AddPrioLines(text string) {
	log.Printf("TextScroller.AddPrioLines: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.prioLines) == 0 {
		t.prioLinesTick = 0
	}

	for _, line := range strings.Split(strings.TrimRight(text, "\n "), "\n") {
		line = strings.TrimRight(line, "\n ")
		line += strings.Repeat(" ", t.Width)
		t.prioLines = append(t.prioLines, line[:t.Width])
		log.Printf("TextScroller.AddPrioLines:  + %v", line)
	}

	if len(t.prioLines)%2 == 1 {
		t.prioLines = append(t.prioLines, strings.Repeat(" ", t.Width))
	}
}

// Tick reformat (scroll) long messages and return formated text
func (t *TextScroller) Tick() (res string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.prioLines) == 0 {
		return t.stdTick()
	}

	t.prioLinesTick++
	if t.prioLinesTick > t.PrioMsgTime {
		t.prioLines = t.prioLines[2:]
		t.prioLinesTick = 0
	}

	if len(t.prioLines) == 0 {
		return t.stdTick()
	}

	res = t.prioLines[0] + "\n" + t.prioLines[1]
	log.Printf("TextScroller.Tick prio '%v'", res)
	return
}

func (t *TextScroller) stdTick() (res string) {
	for _, l := range t.lines {
		res += l.scroll()[:t.Width] + "\n"
	}
	return strings.TrimRight(res, "\n")

}
