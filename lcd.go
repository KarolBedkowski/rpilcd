package main

import (
	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	// Timing constants
	ePulse = 50 * time.Microsecond
	eDelay = 50 * time.Microsecond

	lcdRS = 7
	lcdE  = 8
	lcdD4 = 25
	lcdD5 = 24
	lcdD6 = 23
	lcdD7 = 18

	// Define some device constants
	lcdWidth = 16 // Maximum characters per line
	lcdChr   = true
	lcdCmd   = false

	lcdLine1 = 0x80 // LCD RAM address for the 1st line
	lcdLine2 = 0xC0 // LCD RAM address for the 2nd line
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
	sync.Mutex

	lcdRS gpio.Pin
	lcdE  gpio.Pin
	lcdD4 gpio.Pin
	lcdD5 gpio.Pin
	lcdD6 gpio.Pin
	lcdD7 gpio.Pin

	line1  string
	line2  string
	active bool

	msg chan string
	end chan bool
}

// NewLcd create and init new lcd output
func NewLcd() (l *Lcd) {
	l = &Lcd{
		lcdRS:  initPin(lcdRS),
		lcdE:   initPin(lcdE),
		lcdD4:  initPin(lcdD4),
		lcdD5:  initPin(lcdD5),
		lcdD6:  initPin(lcdD6),
		lcdD7:  initPin(lcdD7),
		active: true,
		msg:    make(chan string),
		end:    make(chan bool),
	}
	l.reset()

	go func() {
		for {
			select {
			case msg := <-l.msg:
				l.display(msg)
			case _ = <-l.end:
				l.close()
				break
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
	if l.active {
		l.end <- true
	}
}

func initPin(pin int) (p gpio.Pin) {
	var err error
	p, err = rpi.OpenPin(pin, gpio.ModeOutput)
	if err != nil {
		panic(err)
	}
	return
}

func (l *Lcd) reset() {
	log.Printf("Lcd.reset()")
	l.writeByte(0x33, lcdCmd) // 110011 Initialise
	l.writeByte(0x32, lcdCmd) // 110010 Initialise
	l.writeByte(0x28, lcdCmd) // 101000 Data length, number of lines, font size
	l.writeByte(0x0C, lcdCmd) // 001100 Display On,Cursor Off, Blink Off
	l.writeByte(0x06, lcdCmd) // 000110 Cursor move direction
	l.writeByte(0x01, lcdCmd) // 000001 Clear display
}

func (l *Lcd) close() {
	l.Lock()
	defer l.Unlock()

	if !l.active {
		return
	}

	l.reset()

	l.lcdRS.Clear()
	l.lcdRS.Close()
	l.lcdE.Clear()
	l.lcdE.Close()
	l.lcdD4.Clear()
	l.lcdD4.Close()
	l.lcdD5.Clear()
	l.lcdD5.Close()
	l.lcdD6.Clear()
	l.lcdD6.Close()
	l.lcdD7.Clear()
	l.lcdD7.Close()

	l.active = false
	close(l.msg)
	close(l.end)
}

// writeByte send byte to lcd
func (l *Lcd) writeByte(bits uint8, characterMode bool) {
	if characterMode {
		l.lcdRS.Set()
	} else {
		l.lcdRS.Clear()
	}

	// High bits
	if bits&0x10 == 0x10 {
		l.lcdD4.Set()
	} else {
		l.lcdD4.Clear()
	}
	if bits&0x20 == 0x20 {
		l.lcdD5.Set()
	} else {
		l.lcdD5.Clear()
	}
	if bits&0x40 == 0x40 {
		l.lcdD6.Set()
	} else {
		l.lcdD6.Clear()
	}
	if bits&0x80 == 0x80 {
		l.lcdD7.Set()
	} else {
		l.lcdD7.Clear()
	}

	// Toggle 'Enable' pin
	time.Sleep(eDelay)
	l.lcdE.Set()
	time.Sleep(ePulse)
	l.lcdE.Clear()
	time.Sleep(eDelay)

	// Low bits
	if bits&0x01 == 0x01 {
		l.lcdD4.Set()
	} else {
		l.lcdD4.Clear()
	}
	if bits&0x02 == 0x02 {
		l.lcdD5.Set()
	} else {
		l.lcdD5.Clear()
	}
	if bits&0x04 == 0x04 {
		l.lcdD6.Set()
	} else {
		l.lcdD6.Clear()
	}
	if bits&0x08 == 0x08 {
		l.lcdD7.Set()
	} else {
		l.lcdD7.Clear()
	}
	// Toggle 'Enable' pin
	time.Sleep(eDelay)
	l.lcdE.Set()
	time.Sleep(ePulse)
	l.lcdE.Clear()
	time.Sleep(eDelay)
}

func (l *Lcd) display(msg string) {
	l.Lock()
	defer l.Unlock()

	if !l.active {
		return
	}

	for line, m := range strings.Split(msg, "\n") {
		//m = removeNlChars(m)
		if len(m) < lcdWidth {
			m = m + strings.Repeat(" ", lcdWidth-len(m))
		}

		switch line {
		case 0:
			if l.line1 == m {
				continue
			}
			l.line1 = m
			l.writeByte(lcdLine1, lcdCmd)
		case 1:
			if l.line2 == m {
				continue
			}
			l.line2 = m
			l.writeByte(lcdLine2, lcdCmd)
		default:
			return
		}

		//log.Printf("Lcd.LcdString Line: %d, msg=%v\n", line, m)
		for i := 0; i < lcdWidth; i++ {
			l.writeByte(m[i], lcdChr)
		}
	}
}
