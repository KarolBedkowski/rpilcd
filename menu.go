package main

import (
	"github.com/naoina/toml"
	"io/ioutil"
	"log"
	"os"
)

type Menu struct {
	Name  string
	Kind  string
	Cmd   string
	Items []*Menu
}

type MainMenu struct {
	Menu Menu
}

var mainMenu MainMenu

func loadMenu(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(buf, &mainMenu); err != nil {
		return err
	}
	return nil
}

func (m *Menu) GetItems(rows, offset, cursor int) (res []string) {
	if offset >= len(m.Items)-rows {
		offset = len(m.Items) - rows
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
		if i == cursor {
			res = append(res, "->"+m.Items[i].Name)
		} else {
			res = append(res, "  "+m.Items[i].Name)
		}
	}
	return
}
