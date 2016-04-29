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
	hasCursor  bool
}

func (tsl *textScrollerLine) set(inp string, width int) {
	if inp == tsl.lineOrg {
		return
	}
	log.Printf("textScrollerLine.set : %+v", inp)
	tsl.hasCursor = len(inp) > 0 && inp[0] == CharCursor[0]
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
		if tsl.hasCursor {
			tsl.line = string(tsl.line[0]) + tsl.line[2:] + string(tsl.line[1])
		} else {
			tsl.line = tsl.line[1:] + string(tsl.line[0])
		}
	}
	return tsl.line
}

// TextScroller format some text to display in few character display
type TextScroller struct {
	Width  int
	Height int
	lines  []*textScrollerLine

	mu sync.Mutex
}

// NewTextScroller create new TextScroller struct
func NewTextScroller(width, height int) *TextScroller {
	res := &TextScroller{
		Width:  width,
		Height: height,
	}
	for i := 0; i < height; i++ {
		l := &textScrollerLine{}
		l.set(strings.Repeat(" ", width), width)
		res.lines = append(res.lines, l)
	}
	return res
}

// Set put some text to TextScroller
func (t *TextScroller) Set(text string) {
	//log.Printf("TextScroller.Set: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	for line, l := range strings.SplitN(text, "\n", t.Height) {
		t.lines[line].set(l, t.Width)
	}
}

// Tick reformat (scroll) long messages and return formated text
func (t *TextScroller) Tick() (res string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, l := range t.lines {
		res += l.scroll()[:t.Width] + "\n"
	}

	return strings.TrimRight(res, "\n")
}
