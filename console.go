package main

import (
	"log"
	"strings"
)

type Console struct {
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
				for i, l := range strings.Split(msg, "\n") {
					log.Printf("SimpleDisplay: [%d] '%s'", i, l)
				}

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
