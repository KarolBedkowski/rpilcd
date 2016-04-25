package main

import (
	"log"
	"os/exec"
)

type Task struct {
	Kind string
	Cmd  string
	Args []string
}

func (r *Task) Execute() (string, bool) {
	out, err := exec.Command(r.Cmd, r.Args...).CombinedOutput()
	res := string(out)
	log.Printf("Execute: err=%v, res=%v", err, res)
	return res, err == nil
}
