package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Merovius/systemd"
	"github.com/prometheus/client_golang/prometheus"
)

const lcdWidth = 16
const lcdHeight = 2
const minMpdActInterval = time.Duration(100) * time.Millisecond

// AppVersion global var
var AppVersion = "dev"

// AppDate build date
var AppDate = "unknown"

// Display (output) interface
type Display interface {
	Display(string)
	Close()
	ToggleBacklight()
	Active() bool
}

func main() {
	//go func() {
	//logger.Println(http.ListenAndServe(":6060", nil))
	//}()
	soutput := flag.Bool("console", false, "Print on console instead of lcd")
	lcdOffOnStart := flag.Bool("off-on-start", false, "Turn off lcd on start")
	logLevel := flag.Int("log-level", 1, "Log level (3=debug, 2=info, 1=error, 0=silent)")
	flag.Parse()

	logger.SetLogLevel(*logLevel)

	systemd.NotifyStatus("starting")
	systemd.AutoWatchdog()

	logger.Infof("RPI LCD ver %s (build %s) starting...", AppVersion, AppDate)

	err := loadConfiguration()
	if err != nil {
		panic(err)
	}
	logger.Debugf("configuration: %#v", configuration)

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
		logger.Infof("webserver starting (%s)...", configuration.ServicesConf.HTTPServerAddr)
		go http.ListenAndServe(configuration.ServicesConf.HTTPServerAddr, nil)
	}

	defer func() {
		if e := recover(); e != nil {
			logger.Infof("Recover: %v", e)
		}
		systemd.Notify("STOPPING=1\r\nSTATUS=stopping")
		logger.Info("main.defer: closing lirc")
		lirc.Close()
		logger.Info("main.defer: closing disp")
		scrMgr.Close()
		logger.Info("main.defer: closing mpd")
		mpd.Close()
		time.Sleep(2 * time.Second)
		logger.Info("main.defer: all closed")
		systemd.NotifyStatus("stopped")
	}()

	mpd.Connect()
	scrMgr.UpdateMpdStatus(MPDGetStatus())
	scrMgr.display(false)

	time.Sleep(1 * time.Second)

	logger.Debugln("main: entering loop")

	ticker := createTicker()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)

	sigHup := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	if *lcdOffOnStart {
		scrMgr.NewCommand(configuration.Keys.ToggleLCD)
	}

	systemd.NotifyReady()
	systemd.NotifyStatus("running")

	for {
		select {
		case _ = <-sig:
			return
		case _ = <-sigHup:
			logger.Info("Reloading configuration")
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
			msg.Free()
		case <-ticker.C:
			scrMgr.Tick()
		}
	}
}

func createTicker() *time.Ticker {
	return time.NewTicker(time.Duration(configuration.DisplayConf.RefreshInterval) * time.Millisecond)
}
