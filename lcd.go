package main

import (
	"github.com/davecheney/gpio"
	_ "github.com/davecheney/gpio/rpi"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	//	"log"
	"strings"
	"sync"
	"time"
)

const (
	// Timing constants
	E_PULSE = 20 * time.Microsecond
	E_DELAY = 20 * time.Microsecond

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

	mu sync.Mutex
}

func initPin(pin int) gpio.Pin {
	if pin, err := gpio.OpenPin(pin, gpio.ModeOutput); err == nil {
		return pin
	} else {
		panic(err)
	}
	return nil
}

func InitLcd() (l *Lcd) {
	l = &Lcd{
		lcdRS: initPin(LCD_RS),
		lcdE:  initPin(LCD_E),
		lcdD4: initPin(LCD_D4),
		lcdD5: initPin(LCD_D5),
		lcdD6: initPin(LCD_D6),
		lcdD7: initPin(LCD_D7),
	}
	//	l.mu.Lock()
	//	defer l.mu.Unlock()

	l.reset()
	return l
}

func (l *Lcd) reset() {
	l.writeByte(0x33, LCD_CMD)
	l.writeByte(0x32, LCD_CMD)
	l.writeByte(0x28, LCD_CMD)
	l.writeByte(0x0C, LCD_CMD)
	l.writeByte(0x06, LCD_CMD)
	l.writeByte(0x01, LCD_CMD)
}

func (l *Lcd) Close() {
	//	l.mu.Lock()
	//	defer l.mu.Unlock()

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

func (l *Lcd) Display(msg string) {
	//	l.mu.Lock()
	//	defer l.mu.Unlock()

	for line, m := range strings.Split(msg, "\n") {
		switch line {
		case 0:
			l.writeByte(LCD_LINE_1, LCD_CMD)
		case 1:
			l.writeByte(LCD_LINE_2, LCD_CMD)
		default:
			return
		}

		//m = removeNlChars(m)

		if len(m) < LCD_WIDTH {
			m = m + strings.Repeat(" ", LCD_WIDTH-len(m))
		}
		//log.Printf("Lcd.LcdString Line: %d, msg=%v\n", line, m)
		for i := 0; i < LCD_WIDTH; i++ {
			l.writeByte(m[i], LCD_CHR)
		}
	}
}
