package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"time"
)

type ButtonEvent struct {
	Pressed bool
	Held    bool
}

func (b ButtonEvent) String() string {
	action := "pressed"
	if !b.Pressed {
		action = "released"
	}
	if b.Held {
		action = "held"
	}
	return fmt.Sprintf("Button was %v", action)
}

// InitButtons initializes all the button pins and fetches a button event channel
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
				Held:    false,
			}
		}
	}
}
