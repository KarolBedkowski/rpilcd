package main

import (
	"github.com/naoina/toml"
	"io/ioutil"
	"log"
	"os"
)

type (
	Menu struct {
		Name  string
		Kind  string
		Cmd   string
		Args  []string
		Items []*Menu
	}
	Keys struct {
		Menu struct {
			Show   string
			Back   string
			Up     string
			Down   string
			Select string
		}
		MPD struct {
			Play    string
			Stop    string
			Pause   string
			Next    string
			Prev    string
			VolUp   string
			VolDown string
		}
	}
)

type Configuration struct {
	Menu Menu
	Keys Keys
}

var configuration Configuration

func loadConfiguration(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(buf, &configuration); err != nil {
		return err
	}
	return nil
}

func (m *Menu) GetItems(rows, offset, cursor int) (res []string) {
	if offset >= len(m.Items)-rows {
		offset = len(m.Items) - rows
	}
	if offset < 0 {
		offset = 0
	}
	if cursor >= len(m.Items) {
		cursor = len(m.Items) - 1
	}

	lastPos := rows + offset
	if lastPos > len(m.Items) {
		lastPos = len(m.Items)
	}

	log.Printf("row=%d, offset=%d, cursor=%d, lastPos=%d", rows, offset, cursor, lastPos)

	for i := offset; i < lastPos; i++ {
		log.Printf("Menu %d: %v", i, m.Items[i])
		if i == cursor {
			res = append(res, "->"+m.Items[i].Name)
		} else {
			res = append(res, "  "+m.Items[i].Name)
		}
	}
	for len(res) < rows {
		res = append(res, "")
	}
	return
}
