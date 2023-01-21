//go:build !pi

package button

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// InitButton initializes all the button pins and fetches a button event channel
func InitButton() <-chan Event {
	log.Infoln("Initializing button handler")

	c := make(chan Event, 5)
	go simulateButton(c)
	return c
}

func simulateButton(c chan<- Event) {
	hupChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	defer close(hupChan)

	for {
		<-hupChan
		c <- Event{
			Pressed: true,
		}
	}
}
