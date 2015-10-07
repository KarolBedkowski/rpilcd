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
	E_PULSE = 50 * time.Microsecond
	E_DELAY = 50 * time.Microsecond

	LCD_RS = 7
	LCD_E  = 8
	LCD_D4 = 25
	LCD_D5 = 24
	LCD_D6 = 23
	LCD_D7 = 18

	// Define some device constants
	LCD_WIDTH = 16 // Maximum characters per line
	LCD_CHR   = true
	LCD_CMD   = false

	LCD_LINE_1 = 0x80 // LCD RAM address for the 1st line
	LCD_LINE_2 = 0xC0 // LCD RAM address for the 2nd line
)

func removeNlChars(str string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	str, _, _ = transform.String(t, str)
	return str
}

type Lcd struct {
	lcdRS gpio.Pin
	lcdE  gpio.Pin
	lcdD4 gpio.Pin
	lcdD5 gpio.Pin
	lcdD6 gpio.Pin
	lcdD7 gpio.Pin

	line1  string
	line2  string
	active bool

	sync.Mutex

	msg chan string
	end chan bool
}

func NewLcd() (l *Lcd) {
	l = initLcd()
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

func (l *Lcd) Display(msg string) {
	l.msg <- msg
}

func (l *Lcd) Close() {
	log.Printf("Lcd.Close")
	l.end <- true
}

func initPin(pin int) gpio.Pin {
	if pin, err := rpi.OpenPin(pin, gpio.ModeOutput); err == nil {
		return pin
	} else {
		panic(err)
	}
	return nil
}

func initLcd() (l *Lcd) {
	l = &Lcd{
		lcdRS:  initPin(LCD_RS),
		lcdE:   initPin(LCD_E),
		lcdD4:  initPin(LCD_D4),
		lcdD5:  initPin(LCD_D5),
		lcdD6:  initPin(LCD_D6),
		lcdD7:  initPin(LCD_D7),
		active: true,
		msg:    make(chan string),
		end:    make(chan bool),
	}
	l.Lock()
	defer l.Unlock()

	l.reset()
	return l
}

func (l *Lcd) reset() {
	log.Printf("Lcd.reset()")
	l.writeByte(0x33, LCD_CMD) // 110011 Initialise
	l.writeByte(0x32, LCD_CMD) // 110010 Initialise
	l.writeByte(0x28, LCD_CMD) // 101000 Data length, number of lines, font size
	l.writeByte(0x0C, LCD_CMD) // 001100 Display On,Cursor Off, Blink Off
	l.writeByte(0x06, LCD_CMD) // 000110 Cursor move direction
	l.writeByte(0x01, LCD_CMD) // 000001 Clear display
	time.Sleep(E_DELAY)
	l.writeByte(0x01, LCD_CMD) // 000001 Clear display
	time.Sleep(E_DELAY)
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
	time.Sleep(E_DELAY)
	l.lcdE.Set()
	time.Sleep(E_PULSE)
	l.lcdE.Clear()
	time.Sleep(E_DELAY)

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
	time.Sleep(E_DELAY)
	l.lcdE.Set()
	time.Sleep(E_PULSE)
	l.lcdE.Clear()
	time.Sleep(E_DELAY)
}

func (l *Lcd) display(msg string) {
	l.Lock()
	defer l.Unlock()

	if !l.active {
		return
	}

	for line, m := range strings.Split(msg, "\n") {
		//m = removeNlChars(m)
		if len(m) < LCD_WIDTH {
			m = m + strings.Repeat(" ", LCD_WIDTH-len(m))
		}

		switch line {
		case 0:
			if l.line1 == m {
				continue
			}
			l.line1 = m
			l.writeByte(LCD_LINE_1, LCD_CMD)
		case 1:
			if l.line2 == m {
				continue
			}
			l.line2 = m
			l.writeByte(LCD_LINE_2, LCD_CMD)
		default:
			return
		}

		//log.Printf("Lcd.LcdString Line: %d, msg=%v\n", line, m)
		for i := 0; i < LCD_WIDTH; i++ {
			l.writeByte(m[i], LCD_CHR)
		}
	}
}
