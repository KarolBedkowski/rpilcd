package main

import (
	"flag"
	"github.com/naoina/toml"
	"io/ioutil"
	"os"
)

var confFileName = flag.String("conf", "conf.toml", "Configuration file name")

type (
	KeysConf struct {
		ToggleLCD string `toml:"toggle_lcd"`

		Menu struct {
			Show   string
			Back   string
			Up     string
			Down   string
			Select string
			Up10   string
			Down10 string
		}
		MPD struct {
			Play    string
			Stop    string
			Pause   string
			Next    string
			Prev    string
			VolUp   string
			VolDown string
			VolMute string
		}
	}

	MPDConf struct {
		Host string
	}

	DisplayConf struct {
		RefreshInterval int
		I2CAddr         byte `toml:"i2c_addr"`
	}

	ServicesConf struct {
		HTTPServerAddr string `toml:"http_server_addr"`
		TCPServerAddr  string `toml:"tcp_server_addr"`
	}

	LircConf struct {
		PidFile string
		Remote  string
	}
)

type Configuration struct {
	Menu         *MenuItem
	Keys         KeysConf
	MPDConf      MPDConf      `toml:"mpd"`
	DisplayConf  DisplayConf  `toml:"display"`
	ServicesConf ServicesConf `toml:"services"`
	LircConf     LircConf     `toml:"lirc"`
}

var configuration *Configuration

func loadConfiguration() error {
	f, err := os.Open(*confFileName)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	conf := &Configuration{}
	if err := toml.Unmarshal(buf, conf); err != nil {
		return err
	}
	configuration = conf
	return nil
}
