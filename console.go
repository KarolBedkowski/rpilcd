package main

import (
	"log"
	"strings"
	"sync"
	"time"
)

const (
	consoleDelay = (E_DELAY*4 + E_PULSE*2) * LCD_WIDTH
)

type Console struct {
	sync.Mutex

	msg chan string
	end chan bool
}

func NewConsole() (l *Console) {
	l = &Console{
		msg: make(chan string),
		end: make(chan bool),
	}
	go func() {
		for {
			select {
			case msg := <-l.msg:
				l.display(msg)

			case _ = <-l.end:
				break
			}
		}
	}()
	return l
}

func (l *Console) Display(msg string) {
	l.msg <- msg
}

func (l *Console) Close() {
	log.Printf("Console close")
	l.end <- true
}

func (l *Console) display(text string) {
	l.Lock()
	defer l.Unlock()
	for i, l := range strings.Split(text, "\n") {
		log.Printf("SimpleDisplay: [%d] '%s'", i, l)
		time.Sleep(consoleDelay)
	}
}
