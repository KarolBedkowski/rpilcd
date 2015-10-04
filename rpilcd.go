package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

var AppVersion = "dev"

func main() {

	lcd := InitLcd()
	lcd.LcdString("\n")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for _ = range c {
			lcd.Close()
			os.Exit(0)
		}
	}()

	for {
		lcd.LcdString(time.Now().String() + "\nTEST\n")
		time.Sleep(1 * time.Second)
	}
}
