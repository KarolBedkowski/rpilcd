package main

// Hitachi HD44780U support library

import (
	"bytes"
	"github.com/golang/glog"
	"github.com/zlowred/embd"
	"github.com/zlowred/embd/controller/hd44780"
	_ "github.com/zlowred/embd/host/rpi"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func removeNlChars(str string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	str, _, _ = transform.String(t, str)
	return str
}

// Lcd output
type Lcd struct {
	hd *hd44780.HD44780
	// max lines
	Lines int
	// LCD width (number of character in line)
	Width int

	// i2c address
	addr      byte
	lastLines [][]byte
	active    bool
	backlight bool
}

// NewLcd create and init new lcd output
func NewLcd() *Lcd {
	var l *Lcd
	defer func() {
		if e := recover(); e != nil {
			glog.Infof("NewLcd failed create - recover: %v", e)
		}
	}()

	l = &Lcd{
		Lines:     2,
		Width:     lcdWidth,
		lastLines: make([][]byte, 2, 2),
	}
	if configuration.DisplayConf.Display == "i2c" {
		l.addr = configuration.DisplayConf.I2CAddr
		glog.V(1).Infof("Starting hd44780 on i2c addr=%d", l.addr)

		if err := embd.InitI2C(); err != nil {
			glog.Error("Can't open lcd: ", err.Error())
			return nil
		}

		bus := embd.NewI2CBus(1)

		var err error
		l.hd, err = hd44780.NewI2C(
			bus,
			l.addr,
			hd44780.PCF8574PinMap,
			hd44780.RowAddress16Col,
			hd44780.TwoLine,
			//		hd44780.BlinkOff,
			hd44780.CursorOff,
			hd44780.EntryIncrement,
		)
		if err != nil {
			glog.Fatal("Can't open i2c lcd: ", err.Error())
			return nil
		}
	} else {
		glog.V(1).Infof("Starting hd44780 on GPIO")
		var err error
		l.hd, err = hd44780.NewGPIO(
			configuration.DisplayConf.GpioRs,
			configuration.DisplayConf.GpioEn,
			configuration.DisplayConf.GpioD4,
			configuration.DisplayConf.GpioD5,
			configuration.DisplayConf.GpioD6,
			configuration.DisplayConf.GpioD7,
			configuration.DisplayConf.GpioBl,
			hd44780.Positive,
			hd44780.RowAddress16Col,
			hd44780.TwoLine,
			//		hd44780.BlinkOff,
			hd44780.CursorOff,
			hd44780.EntryIncrement,
		)
		if err != nil {
			glog.Fatal("Can't open gpio lcd: ", err.Error())
			return nil
		}
	}

	l.hd.Clear()
	l.hd.BacklightOn()
	l.backlight = true

	l.active = true

	// Custom characters
	// play
	l.setChar(0, []byte{0x0, 0x8, 0xc, 0xe, 0xc, 0x8, 0x0, 0x0})
	// pause
	l.setChar(1, []byte{0x0, 0x1b, 0x1b, 0x1b, 0x1b, 0x1b, 0x0, 0x0})
	// stop
	l.setChar(2, []byte{0x0, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x0, 0x0})

	return l
}

// Display show some message
func (l *Lcd) Display(msg string) {
	msgb := []byte(msg)
	for line, text := range bytes.Split(msgb, []byte("\n")) {
		l.DisplayLine(line, text)
	}
}

// DisplayLine display `text` in `line`.
func (l *Lcd) DisplayLine(line int, text []byte) {
	if !l.active || line >= l.Lines || !l.backlight {
		return
	}

	// skip not changed lines
	if bytes.Compare(l.lastLines[line], text) == 0 {
		return
	}

	l.lastLines[line] = text

	textLen := len(text)
	if textLen < lcdWidth {
		text = append(text, bytes.Repeat([]byte(" "), l.Width-textLen)...)
	} else if textLen > l.Width {
		text = text[:l.Width]
	}

	l.hd.SetCursor(0, line)
	for _, c := range text {
		l.hd.WriteChar(byte(c))
	}
}

// Close LCD
func (l *Lcd) Close() {
	glog.Infof("Lcd.Close")
}

// ToggleBacklight turn off/on lcd backlight
func (l *Lcd) ToggleBacklight() {
	if !l.active {
		return
	}
	if l.backlight {
		l.hd.BacklightOff()
		l.hd.Clear()
		l.hd.Home()
		l.backlight = false
	} else {
		l.hd.BacklightOn()
		l.backlight = true
		for i, line := range l.lastLines {
			l.lastLines[i] = []byte{}
			l.DisplayLine(i, line)
		}
	}
}

func (l *Lcd) setChar(pos byte, def []byte) {
	if len(def) != 8 {
		panic("invalid def - req 8 bytes")
	}
	l.hd.WriteInstruction(0x40 + pos*8)
	for _, d := range def {
		l.hd.WriteChar(d)
	}
}

// Active return true when lcd is enabled and working
func (l *Lcd) Active() bool {
	return l.active && l.backlight
}
