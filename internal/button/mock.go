//go:build !pi

package button

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// InitButton initializes all the button pins and fetches a button event channel
func InitButton() <-chan ButtonEvent {
	log.Infoln("Initializing button handler")

	c := make(chan ButtonEvent, 5)
	go simulateButton(c)
	return c
}

func simulateButton(c chan<- ButtonEvent) {
	hupChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	defer close(hupChan)

	for {
		<-hupChan
		c <- ButtonEvent{
			Pressed: true,
		}
	}
}
