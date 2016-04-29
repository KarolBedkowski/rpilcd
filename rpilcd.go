package main

import (
	"flag"
	"log"
	//_ "net/http/pprof"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const lcdWidth = 16
const lcdHeight = 2
const minMpdActInterval = time.Duration(100) * time.Millisecond

// AppVersion global var
var AppVersion = "dev"

// Display (output) interface
type Display interface {
	Display(string)
	Close()
	ToggleBacklight()
}

func main() {

	//go func() {
	//log.Println(http.ListenAndServe(":6060", nil))
	//}()

	log.Printf("RPI LCD ver %s starting...", AppVersion)

	soutput := flag.Bool("console", false, "Print on console instead of lcd")
	flag.Parse()

	err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	log.Printf("configuration: %#v", configuration)

	ws := UMServer{
		Addr: configuration.ServicesConf.TCPServerAddr,
	}
	if configuration.ServicesConf.TCPServerAddr != "" {
		ws.Start()
	}

	if configuration.ServicesConf.HTTPServerAddr != "" {
		http.Handle("/metrics", prometheus.Handler())
		go http.ListenAndServe(configuration.ServicesConf.HTTPServerAddr, nil)
	}

	mpd := NewMPD()
	scrMgr := NewScreenMgr(*soutput)
	lirc := NewLirc()

	defer func() {
		if e := recover(); e != nil {
			log.Printf("Recover: %v", e)
		}
		log.Printf("main.defer: closing disp")
		scrMgr.Close()
		log.Printf("main.defer: closing mpd")
		mpd.Close()
		lirc.Close()
		time.Sleep(2 * time.Second)
		log.Printf("main.defer: all closed")
	}()

	mpd.Connect()
	scrMgr.UpdateMpdStatus(MPDGetStatus())

	time.Sleep(1 * time.Second)

	log.Printf("main: entering loop")

	ticker := createTicker()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	sigHup := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	for {
		select {
		case _ = <-sig:
			return
		case _ = <-sigHup:
			log.Printf("Reloading configuration")
			ticker.Stop()
			err := loadConfiguration()
			if err != nil {
				panic(err)
			}
			ticker = createTicker()
		case ev := <-lirc.Events:
			scrMgr.NewCommand(ev)
		case msg := <-ws.Message:
			scrMgr.NewCommand(msg)
		case msg := <-mpd.Message:
			scrMgr.UpdateMpdStatus(msg)
		case <-ticker.C:
			scrMgr.Tick()
		}
	}
}

func createTicker() *time.Ticker {
	return time.NewTicker(time.Duration(configuration.DisplayConf.RefreshInterval) * time.Millisecond)
}
