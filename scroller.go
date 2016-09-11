package main

import (
	"bytes"
	"github.com/golang/glog"
	"strings"
	"sync"
)

type textScrollerLine struct {
	lineOrg    string
	line       []byte
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
		tsl.line = append([]byte(inp), ' ', '|', ' ')
		return
	}
	b := []byte(inp)
	for len(b) < width {
		b = append(b, ' ')
	}
	tsl.line = b
}

func (tsl *textScrollerLine) getAndScroll() []byte {
	if !tsl.needScroll {
		return tsl.line
	}
	newLine := make([]byte, 0, len(tsl.line))
	if tsl.fixPart > 0 {
		newLine = append(newLine, tsl.line[:tsl.fixPart]...)
		newLine = append(newLine, tsl.line[tsl.fixPart+1:]...)
		newLine = append(newLine, tsl.line[tsl.fixPart])
	} else {
		newLine = append(newLine, tsl.line[1:]...)
		newLine = append(newLine, tsl.line[0])
	}
	tsl.line = newLine
	return newLine
}

// TextScroller format some text to display in few character display
type TextScroller struct {
	Width  int
	Height int
	lines  []*textScrollerLine

	mu sync.Mutex
}

// NewTextScroller create new TextScroller struct
func NewTextScroller(width, height int) TextScroller {
	res := TextScroller{
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
	result := make([]byte, 0, t.Width*2+2)

	for _, l := range t.lines {
		result = append(result, l.getAndScroll()[:t.Width]...)
		result = append(result, '\n')
	}

	result = bytes.TrimRight(result, "\n")
	return string(result)
}

// Get current strings
func (t *TextScroller) Get() (res string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make([]byte, 0, t.Width*2+2)

	for _, l := range t.lines {
		result = append(result, l.line[:t.Width]...)
		result = append(result, '\n')
	}

	result = bytes.TrimRight(result, "\n")
	return string(result)
}
