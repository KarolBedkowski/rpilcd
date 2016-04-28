package main

// Hitachi HD44780U support library

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"k.prv/go-hd44780"
	"log"
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
	msg chan string
	end chan bool

	lcd *hd44780.I2C4bit
}

// NewLcd create and init new lcd output
func NewLcd() (l *Lcd) {
	l = &Lcd{
		msg: make(chan string, 10),
		end: make(chan bool, 1),
		lcd: hd44780.NewI2C4bit(0x3f),
	}

	err := l.lcd.Open()
	if err != nil {
		log.Panic("Can't open lcd: %s", err.Error())
		return nil
	}

	if !l.lcd.Active() {
		log.Panic("LCD interface is inactive!")
		return nil
	}

	// Custom characters
	// play
	l.lcd.SetChar(0, []byte{0x0, 0x8, 0xc, 0xe, 0xc, 0x8, 0x0, 0x0})
	// pause
	l.lcd.SetChar(1, []byte{0x0, 0x1b, 0x1b, 0x1b, 0x1b, 0x1b, 0x0, 0x0})
	// stop
	l.lcd.SetChar(2, []byte{0x0, 0x1f, 0x1f, 0x1f, 0x1f, 0x1f, 0x0, 0x0})

	go func() {
		for {
			select {
			case msg := <-l.msg:
				l.lcd.DisplayLines(msg)
			case _ = <-l.end:
				l.lcd.Close()
				return
			}
		}
	}()
	return l
}

// Display show some message
func (l *Lcd) Display(msg string) {
	l.msg <- msg
}

// Close LCD
func (l *Lcd) Close() {
	log.Printf("Lcd.Close")
	l.end <- true
}

func (l *Lcd) ToggleBacklight() {
	l.lcd.ToggleBacklight()
}
