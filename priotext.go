package main

import (
	"log"
	"strings"
	"sync"
)

// PrioText keep text for display as priority
type PrioText struct {
	mu sync.Mutex

	Width         int
	prioLines     []string
	prioLinesTick int
	PrioMsgTime   int
}

// NewPrioText create new PrioText struct
func NewPrioText(width int) *PrioText {
	res := &PrioText{
		Width:     width,
		prioLines: make([]string, 0),
	}
	return res
}

// Add text to immediate show in display
func (t *PrioText) Add(text string) {
	log.Printf("PrioText.add: %v", text)
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.prioLines) == 0 {
		t.prioLinesTick = 0
	}

	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			line += strings.Repeat(" ", t.Width)
			t.prioLines = append(t.prioLines, line[:t.Width])
			log.Printf("PrioText.AddPrioLines:  + %v", line)
		}
	}

	if len(t.prioLines)%2 == 1 {
		t.prioLines = append(t.prioLines, strings.Repeat(" ", t.Width))
	}
}

// Get return formated priority text
func (t *PrioText) Get() (res string, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.prioLines) == 0 {
		return "", false
	}

	t.prioLinesTick++
	if t.prioLinesTick > t.PrioMsgTime {
		t.prioLines = t.prioLines[2:]
		t.prioLinesTick = 0
	}

	if len(t.prioLines) == 0 {
		return "", false
	}

	res = t.prioLines[0] + "\n" + t.prioLines[1]
	log.Printf("PrioText.Get -> '%v'", res)
	return res, true
}
