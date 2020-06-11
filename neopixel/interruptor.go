package neopixel

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Queue struct {
	waiting       int
	runLock       sync.Mutex
	interruptLock sync.Mutex
}

type Unlocker func()

// Wait for your turn on a resource. Will try to interrupt any other user
func (i *Queue) Queue() Unlocker {
	i.interrupt()
	i.runLock.Lock()
	return func() {
		i.done()
	}
}

func (i *Queue) interrupt() {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	i.waiting++
}

func (i *Queue) IsInterrupted() bool {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	return i.waiting != 0
}

func (i *Queue) done() {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()
	defer i.runLock.Unlock()

	i.waiting--
	if i.waiting < 0 {
		log.Warn(errors.New("number waiting in queue less than zero"))
	}
}

