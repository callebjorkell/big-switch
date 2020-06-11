package main

import (
	"fmt"
	"github.com/callebjorkell/big-switch/lcd"
	"github.com/callebjorkell/big-switch/neopixel"
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
	default:
		kingpin.FatalUsage("Unrecognized command")
	}
}

func startServer() {
	lcd.InitLCD()
	lcd.ClearAll()

	c := neopixel.Breathe(0xff3399)

	time.Sleep(2 * time.Second)

	go func() {
		events := InitButton()
		for {
			select {
			case e := <-events:
				log.Infof("Event: %v", e)
				if e.Pressed {
					c.Close()
					neopixel.Flash(0x0000ff)
				}
			}
		}
	}()

	lcd.PrintLine(lcd.Line1, "    Awesome!")
	time.Sleep(2 * time.Second)
	for i := 0; i<16; i++ {
		str := ""
		for j := 0; j<=i; j++ {
			str = str + "*"
		}
		lcd.PrintLine(lcd.Line2, str)
		time.Sleep(1 * time.Second)
	}

	c.Close()

	lcd.PrintLine(lcd.Line1, "   Good bye...")
	lcd.Clear(lcd.Line2)
	log.Info("Done...")
}