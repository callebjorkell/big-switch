package main

import (
	"fmt"
	"github.com/callebjorkell/big-switch/internal/button"
	"github.com/callebjorkell/big-switch/internal/deploy"
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

func startServer() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	lcd.InitLCD()
	lcd.Reset()

	led := neopixel.NewLedController()

	conf, err := readConfig()
	if err != nil {
		panic(err)
	}

	checker := deploy.NewChecker(conf.ReleaseManager.Token)

	for _, service := range conf.Services {
		checker.AddWatch(service.Name)
	}

	confirm := make(chan bool)
	defer close(confirm)
	go func() {
		events := button.InitButton()
		for {
			select {
			case e := <-events:
				log.Infof("Event: %v", e)
				if e.Pressed {
					// non-blocking confirm. If the button is not armed, we do not care.
					select {
					case confirm <- true:
					default:
					}
				}
			}
		}
	}()

	go func() {
		colors := make(map[string]uint32)
		for _, service := range conf.Services {
			colors[service.Name] = service.Color
		}

		events := checker.Changes()
		for {
			select {
			case e := <-events:
				log.Infof("Service %s changed. Waiting for confirmation!", e.Service)
				led.Breathe(colors[e.Service])
				lcd.PrintLine(lcd.Line1, "Press to deploy!")
				lcd.PrintLine(lcd.Line2, e.Service)

				select {
				case confirmed := <-confirm:
					if confirmed {
						// Do actual deploy here.
						err := fmt.Errorf("TODO: Implement me")
						if err != nil {
							log.Warn("Unable to trigger deploy: ", err)
							lcd.PrintLine(lcd.Line1, " TRIGGER FAILED")
							led.Flash(0xff0000)
							<-time.After(5 * time.Second)
						}
						led.Flash(0x00ff00)
					}
				case <-time.After(30 * time.Second):
					log.Info("Confirmation timed out.")
				}
				lcd.Reset()
				led.Stop()
			}
		}
	}()

	select {
	case <-signalChan:
	}

	lcd.PrintLine(lcd.Line1, "  Sleeping...")
	lcd.Clear(lcd.Line2)
	led.Close()

	log.Info("Done...")
}
