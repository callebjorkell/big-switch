package main

import (
	"fmt"
	"github.com/callebjorkell/big-switch/internal/lcd"
	"github.com/callebjorkell/big-switch/internal/neopixel"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	app     = kingpin.New("big-switch", "Big switch trigger")
	debug   = app.Flag("debug", "Turn on debug logging.").Bool()
	start   = app.Command("start", "Start the deployer")
	version = app.Command("version", "Show current version.")
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-signalChan:
			os.Exit(0)
		}
	}()

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("%v: Try --help\n", err.Error())
		os.Exit(1)
	}

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	if *debug {
		log.Info("Enabling debug output...")
		log.SetLevel(log.DebugLevel)
	}

	switch cmd {
	case start.FullCommand():
		startServer()
	case version.FullCommand():
		showVersion()
	default:
		kingpin.FatalUsage("Unrecognized command")
	}
}

var buildTime, buildVersion string

func showVersion() {
	if buildTime != "" && buildVersion != "" {
		fmt.Printf("%s (built: %s)\n", buildVersion, buildTime)
	} else {
		fmt.Println("nfc-player: dev")
	}
}

func startServer() {
	lcd.InitLCD()

	lcd.ClearAll()

	time.Sleep(2 * time.Second)

	lcd.PrintLine(lcd.Line1, "   Awesome!")
	time.Sleep(2 * time.Second)
	for i := 0; i < 16; i++ {
		str := ""
		for j := 0; j <= i; j++ {
			str = str + "*"
		}
		lcd.PrintLine(lcd.Line2, str)
		time.Sleep(1 * time.Second)
	}

	neopixel.Test()

	<-time.After(5 * time.Second)
}
