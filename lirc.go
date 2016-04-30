package main

import (
	"github.com/chbmuc/lirc"
	"github.com/golang/glog"
)

type Lirc struct {
	ir     *lirc.Router
	Events chan string
}

func NewLirc() *Lirc {
	l := &Lirc{
		Events: make(chan string, 10),
	}
	if configuration.LircConf.PidFile == "" {
		glog.Error("Lirc not configured")
		return l
	}

	var err error
	l.ir, err = lirc.Init(configuration.LircConf.PidFile)
	if err != nil {
		return l
	}

	remote := configuration.LircConf.Remote
	if remote == "" {
		remote = "*"
	}

	l.ir.Handle(remote, "*", l.handler)

	go l.ir.Run()

	return l
}

func (l *Lirc) handler(event lirc.Event) {
	if glog.V(1) {
		glog.Infof("lirc ir event: %#v", event)
	}
	l.Events <- event.Button
}

func (l *Lirc) Close() {
	if l.ir != nil {
		l.ir.Close()
	}
}
