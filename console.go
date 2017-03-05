package main

import (
	"strings"
	"time"
)

const (
	consoleDelay = (80 * time.Millisecond) // * 16 * 2
)

// Console simulate lcd without physical lcd
type Console struct {
	active bool
}

// NewConsole create and init new console output
func NewConsole() (l *Console) {
	l = &Console{
		active: true,
	}
	return l
}

// Display show some message
func (l *Console) Display(msg string) {
	l.display(msg)
}

// Close console
func (l *Console) Close() {
	logger.Infof("Console close")
	l.active = false
}

func (l *Console) display(text string) {
	if l.active {
		for i, l := range strings.Split(text, "\n") {
			logger.Printf("SimpleDisplay: [%d] '%s'", i, l)
			time.Sleep(consoleDelay)
		}
	} else {
		logger.Infof("SimpleDisplay: not active")
	}
}

func (l *Console) ToggleBacklight() {
	logger.Printf("SimpleDisplay: toggle backlight")
	l.active = !l.active
}

func (l *Console) Active() bool {
	return l.active
}
