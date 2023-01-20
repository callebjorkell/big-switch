//go:build pi

package button

import (
	log "github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"time"
)

// InitButton initializes all the button pins and fetches a button event channel
func InitButton() <-chan ButtonEvent {
	log.Infoln("Initializing button handler")
	button := gpioreg.ByName("GPIO20")

	c := make(chan ButtonEvent, 5)
	go handleButton(button, c)
	return c
}

func handleButton(b gpio.PinIO, c chan ButtonEvent) {
	if err := b.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}

	last := b.Read()
	for {
		// wait for the edge
		if !b.WaitForEdge(time.Second) {
			continue
		}

		// debounce
		l := b.Read()
		if l == last {
			continue
		}

		time.Sleep(15 * time.Millisecond)
		if l == b.Read() {
			// ... and handle
			last = l
			c <- ButtonEvent{
				Pressed: l == gpio.Low,
			}
		}
	}
}
