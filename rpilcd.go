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
	refreshInt := flag.Int64("interval", 1000, "Interval between lcd updates in ms")
	satartService := flag.Bool("startService", true, "Start TCP server for urgent messages")
	serviceAddr := flag.String("serviceAddr", "localhost:8681", "TCP server address")
	httpServAddr := flag.String("listen-address", ":8001", "The address to listen on for HTTP requests.")

	flag.Parse()

	ws := UMServer{
		Addr: *serviceAddr,
	}
	if *satartService {
		ws.Start()
	}

	if *httpServAddr != "" {
		http.Handle("/metrics", prometheus.Handler())
		go http.ListenAndServe(*httpServAddr, nil)
	}

	err := loadConfiguration("conf.toml")
	log.Printf("Load menu: %s", err)
	log.Printf("menu: %s", configuration)

	log.Printf("main: interval: %d ms", *refreshInt)

	mpd := NewMPD()
	dispP := NewDisplayP(*soutput, int(*refreshInt), mpd)

	lirc := NewLirc()

	defer func() {
		//if e := recover(); e != nil {
		//	log.Printf("Recover: %v", e)
		//}
		log.Printf("main.defer: closing disp")
		dispP.Close()
		log.Printf("main.defer: closing mpd")
		mpd.Close()
		lirc.Close()
		time.Sleep(2 * time.Second)
		log.Printf("main.defer: all closed")
	}()

	mpd.Connect()
	dispP.UpdateMpdStatus(mpd.GetStatus())

	time.Sleep(1 * time.Second)

	log.Printf("main: entering loop")

	ticker := time.NewTicker(time.Duration(*refreshInt) * time.Millisecond)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		// TODO: handle lirc event
		select {
		case _ = <-sig:
			return
		case ev := <-lirc.Events:
			{
				log.Printf("lirc event: %s", ev)
				dispP.NewCommand(ev)
			}
		case msg := <-ws.Message:
			dispP.NewCommand(msg)
		case msg := <-mpd.Message:
			dispP.UpdateMpdStatus(msg)
		case <-ticker.C:
			dispP.Tick()
		}
	}
}
