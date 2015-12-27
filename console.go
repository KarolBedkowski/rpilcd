package main

import (
	"log"
	"strings"
	"sync"
	"time"
)

const (
	consoleDelay = (eDelay*4 + ePulse*2) * lcdWidth
)

// Console simulate lcd without physical lcd
type Console struct {
	sync.Mutex

	active bool

	msg chan string
	end chan bool
}

// NewConsole create and init new console output
func NewConsole() (l *Console) {
	l = &Console{
		active: true,
		msg:    make(chan string),
		end:    make(chan bool),
	}
	go func() {
		for {
			select {
			case msg := <-l.msg:
				l.display(msg)

			case _ = <-l.end:
				l.close()
				return
			}
		}
	}()
	return l
}

// Display show some message
func (l *Console) Display(msg string) {
	l.msg <- msg
}

// Close console
func (l *Console) Close() {
	log.Printf("Console close")
	if l.active {
		l.end <- true
	}
}

func (l *Console) close() {
	l.Lock()
	defer l.Unlock()

	if !l.active {
		return
	}

	l.active = false
	close(l.msg)
	close(l.end)
}

func (l *Console) display(text string) {
	l.Lock()
	defer l.Unlock()
	for i, l := range strings.Split(text, "\n") {
		log.Printf("SimpleDisplay: [%d] '%s'", i, l)
		time.Sleep(consoleDelay)
	}
}
