package neopixel

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Interruptor is intended to be used for sharing the LEDs between concurrent operations, where for example some animations
// can take a long time. When goproc wants to take control over the LEDs, the request should be queued, which
// will set an "interrupted" state on the interruptor that a running animation can check. If the current resource owner
// sees an interruption on the interruptor, it SHOULD release the resource and let the queued process continue.
type Interruptor struct {
	waiting       int
	runLock       sync.Mutex
	interruptLock sync.Mutex
}

type Unlocker func()

// Interrupt and wait for turn on a resource. Set the interrupted state and then wait for the run lock.
func (i *Interruptor) Interrupt() Unlocker {
	// interrupt and wait to run
	i.interrupt()
	i.runLock.Lock()

	// mark as running, and return a done callback to the caller for unlocking the run lock.
	i.running()
	return func() {
		i.done()
	}
}

func (i *Interruptor) running() {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	i.waiting--
}

func (i *Interruptor) interrupt() {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	i.waiting++
	log.Debug("Added to interruptor: ", i.waiting)
}

func (i *Interruptor) IsInterrupted() bool {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	return i.waiting != 0
}

func (i *Interruptor) done() {
	defer i.runLock.Unlock()

	log.Debug("Marked done. Currently waiting: ", i.waiting)
	if i.waiting < 0 {
		log.Warn(errors.New("number waiting in interruptor less than zero"))
	}
}
