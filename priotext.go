package main

import (
	"encoding/json"
	"io"
	"log"
	"strings"
	"sync"
)

// UrgentMsgManager keep msg for display as priority
type UrgentMsgManager struct {
	mu sync.Mutex

	DefaultTimeout int

	messages []UrgentText
}

// UrgentText to display
type UrgentText struct {
	Text     string
	Timeout  int
	PutOnTop bool
}

// NewUrgentText create UrgentText structure for given msg & timeout and
// default other params
func NewUrgentText(msg string, timeout int) UrgentText {
	return UrgentText{msg, timeout, false}
}

// NewUrgentMsgManager create new UrgentMsgManager struct
func NewUrgentMsgManager(width int) *UrgentMsgManager {
	res := &UrgentMsgManager{
		messages: make([]UrgentText, 0),
	}
	return res
}

// Output return UrgentText structure formatted to display
func (t *UrgentText) Output() string {
	lines := strings.Split(t.Text, "\n")
	if len(lines) == 1 {
		return lines[0] + "\n"
	} else if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[0]) + "\n" + strings.TrimSpace(lines[1])
}

// Add msg to immediate show in display
func (t *UrgentMsgManager) Add(msg string) {
	t.AddWithTime(msg, t.DefaultTimeout)
}

// AddJSON push UrgentText structure in json into queue
func (t *UrgentMsgManager) AddJSON(text string) {
	dec := json.NewDecoder(strings.NewReader(text))

	t.mu.Lock()
	defer t.mu.Unlock()

	for {
		var msg UrgentText
		if err := dec.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			// error - add simple msg
			log.Printf("AddJSON.decode json error: %s", err.Error())
			t.messages = append(t.messages, NewUrgentText(text, t.DefaultTimeout))
			break
		}
		t.messages = append(t.messages, msg)
		if msg.PutOnTop {
			copy(t.messages[1:], t.messages[0:])
			t.messages[0] = msg
		}
	}

	for no, l := range t.messages {
		log.Printf("LINE: %d, '%v'", no, l)
	}
}

// AddWithTime push msg with given timeout into messages queue
func (t *UrgentMsgManager) AddWithTime(msg string, timeout int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.messages = append(t.messages, NewUrgentText(msg, timeout))

}

// Get return formated priority msg
func (t *UrgentMsgManager) Get() (res string, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.messages) == 0 {
		return "", false
	}

	msg := t.messages[0]
	res = msg.Output()

	// decr display time
	if msg.Timeout > 0 {
		msg.Timeout--
		t.messages[0] = msg
	} else {
		t.messages = t.messages[1:]
	}

	log.Printf("UrgentMsgManager.Get -> '%v'", res)
	return res, true
}
