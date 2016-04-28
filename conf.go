package main

import (
	"github.com/naoina/toml"
	"io/ioutil"
	"os"
)

type (
	Keys struct {
		ToggleLCD string `toml:"toggle_lcd"`

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
	Menu *MenuItem
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
