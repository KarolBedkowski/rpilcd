package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"github.com/coreos/go-systemd/daemon"
	"github.com/Merovius/systemd"
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
	//daemon.SdNotify("STATUS=starting")
	systemd.NotifyStatus("starting")

	systemd.AutoWatchdog()

	//go func() {
	//glog.Println(http.ListenAndServe(":6060", nil))
	//}()
	soutput := flag.Bool("console", false, "Print on console instead of lcd")
	flag.Parse()

	glog.Infof("RPI LCD ver %s starting...", AppVersion)

	err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	if glog.V(1) {
		glog.Infof("configuration: %#v", configuration)
	}

	ws := UMServer{
		Addr: configuration.ServicesConf.TCPServerAddr,
	}
	if configuration.ServicesConf.TCPServerAddr != "" {
		ws.Start()
	}

	mpd := NewMPD()
	scrMgr := NewScreenMgr(*soutput)
	lirc := NewLirc()

	if configuration.ServicesConf.HTTPServerAddr != "" {
		http.Handle("/metrics", prometheus.Handler())
		http.HandleFunc("/", scrMgr.WebHandler)
		glog.Infof("webserver starting (%s)...", configuration.ServicesConf.HTTPServerAddr)
		go http.ListenAndServe(configuration.ServicesConf.HTTPServerAddr, nil)
	}

	defer func() {
		//if e := recover(); e != nil {
		//	glog.Infof("Recover: %v", e)
		//}
		//daemon.SdNotify("STOPPING=1")
		systemd.Notify("STOPPING=1\r\nSTATUS=stopping")
		glog.Infof("main.defer: closing lirc")
		lirc.Close()
		glog.Infof("main.defer: closing disp")
		scrMgr.Close()
		glog.Infof("main.defer: closing mpd")
		mpd.Close()
		time.Sleep(2 * time.Second)
		glog.Infof("main.defer: all closed")
		//daemon.SdNotify("STATUS=stopped")
		systemd.NotifyStatus("stopped")
	}()

	mpd.Connect()
	scrMgr.UpdateMpdStatus(MPDGetStatus())
	scrMgr.display(false)

	time.Sleep(1 * time.Second)

	glog.V(1).Infof("main: entering loop")

	ticker := createTicker()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	sigHup := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	//daemon.SdNotify("READY=1")
	systemd.NotifyReady()
	systemd.NotifyStatus("running")

	for {
		select {
		case _ = <-sig:
			return
		case _ = <-sigHup:
			glog.Infof("Reloading configuration")
			ticker.Stop()
			err := loadConfiguration()
			if err != nil {
				panic(err)
			}
			ticker = createTicker()
		case ev := <-lirc.Events:
			if ev != "" {
				scrMgr.NewCommand(ev)
			}
		case msg := <-ws.Message:
			if msg != "" {
				scrMgr.NewCommand(msg)
			}
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
