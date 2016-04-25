package main

import (
	"github.com/chbmuc/lirc"
	"log"
)

type Lirc struct {
	ir     *lirc.Router
	Events chan string
}

func NewLirc() *Lirc {
	l := &Lirc{
		Events: make(chan string, 10),
	}
	var err error
	l.ir, err = lirc.Init("/var/run/lirc/lircd")
	if err != nil {
		return l
	}

	l.ir.Handle("*", "*", l.handler)

	go l.ir.Run()

	return l
}

func (l *Lirc) handler(event lirc.Event) {
	log.Printf("lirc ir event: %#v", event)
	l.Events <- event.Button
}

func (l *Lirc) Close() {
	l.ir.Close()
}
