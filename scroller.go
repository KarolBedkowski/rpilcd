package main

import (
	"github.com/golang/glog"
	"strings"
	"sync"
)

type textScrollerLine struct {
	lineOrg    string
	line       string
	needScroll bool
	fixPart    int
}

func (tsl *textScrollerLine) set(inp string, width int, fixPart int) {
	if inp == tsl.lineOrg {
		return
	}
	if glog.V(1) {
		glog.Infof("textScrollerLine.set : %+v", inp)
	}
	tsl.fixPart = fixPart
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

func (tsl *textScrollerLine) getAndScroll() string {
	res := tsl.line
	if tsl.needScroll {
		if tsl.fixPart > 0 {
			tsl.line = tsl.line[:tsl.fixPart] + tsl.line[tsl.fixPart+1:] + string(tsl.line[tsl.fixPart])
		} else {
			tsl.line = tsl.line[1:] + string(tsl.line[0])
		}
	}
	return res
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
		l.set(strings.Repeat(" ", width), width, 0)
		res.lines = append(res.lines, l)
	}
	return res
}

// Set put some text to TextScroller
func (t *TextScroller) Set(text string, fixPart int) {
	//log.Printf("TextScroller.Set: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	for line, l := range strings.SplitN(text, "\n", t.Height) {
		t.lines[line].set(l, t.Width, fixPart)
	}
}

// Tick reformat (scroll) long messages and return formated text
func (t *TextScroller) Tick() (res string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, l := range t.lines {
		res += l.getAndScroll()[:t.Width] + "\n"
	}

	return strings.TrimRight(res, "\n")
}

// Get current strings
func (t *TextScroller) Get() (res string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, l := range t.lines {
		res += l.line[:t.Width] + "\n"
	}

	return strings.TrimRight(res, "\n")
}
