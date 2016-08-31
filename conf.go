package main

import (
	"flag"
	"github.com/naoina/toml"
	"io/ioutil"
	"os"
)

var confFileName = flag.String("conf", "conf.toml", "Configuration file name")

type (
	// KeysConf keep mapping between actions and Lirc events
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

	// MPDConf keep configuration parameters related do mpd
	MPDConf struct {
		// Host is mpd server address
		Host string
	}

	// DisplayConf keep hardware configuration for lcd display
	DisplayConf struct {
		RefreshInterval int
		I2CAddr         byte `toml:"i2c_addr"`
		Display         string
		GpioRs          uint8
		GpioEn          uint8
		GpioD4          uint8
		GpioD5          uint8
		GpioD6          uint8
		GpioD7          uint8
		GpioBl          uint8
	}

	// ServicesConf store internal web/tcp servers configuration
	ServicesConf struct {
		HTTPServerAddr string `toml:"http_server_addr"`
		TCPServerAddr  string `toml:"tcp_server_addr"`
	}

	// LircConf store information about Lirc configuration
	LircConf struct {
		PidFile string
		Remote  string
	}
)

// Configuration is top configuration object
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

	defer func() {
		if f != nil {
			f.Close()
		}
	}()

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
