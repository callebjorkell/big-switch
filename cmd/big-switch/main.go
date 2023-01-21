package main

import (
	"fmt"
	"github.com/callebjorkell/big-switch/internal/button"
	"github.com/callebjorkell/big-switch/internal/deploy"
	"github.com/callebjorkell/big-switch/internal/lcd"
	"github.com/callebjorkell/big-switch/internal/neopixel"
	"github.com/callebjorkell/big-switch/internal/passphrase"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type colorFormatter struct {
	log.TextFormatter
}

func (f *colorFormatter) Format(entry *log.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case log.DebugLevel, log.TraceLevel:
		levelColor = 90 // dark grey
	case log.WarnLevel:
		levelColor = 33 // yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		levelColor = 91 // bright red
	default:
		levelColor = 39 // default
	}
	return []byte(fmt.Sprintf("\x1b[%dm%s\x1b[0m\n", levelColor, entry.Message)), nil
}

func main() {
	log.SetFormatter(&colorFormatter{})

	if err := RootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func encryptConfig(f string) {
	log.Infof("Will try to encrypt %v", f)
}

func startServer() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	lcd.InitLCD()
	lcd.Reset()

	led := neopixel.NewLedController()

	p := passphrase.NewServer()
	go p.Listen()

	select {
	case <-signalChan:
		p.Close()
		os.Exit(0)
	case pass := <-p.PassChan():
		log.Infof("got passphrase: %v", pass)
		p.Close()
	}

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
