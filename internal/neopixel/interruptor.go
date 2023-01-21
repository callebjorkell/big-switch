package neopixel

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Queue is intended to be used for sharing the LEDs between concurrent operations, where for example some animations
// can take a long time. When goproc wants to take control over the LEDs, the request should be queued, which
// will set an "interrupted" state on the queue that a running animation can check. If the current resource owner
// sees an interruption on the queue, it SHOULD release the resource and let the queued process continue.
type Queue struct {
	waiting       int
	runLock       sync.Mutex
	interruptLock sync.Mutex
}

type Unlocker func()

// Queue and wait for turn a on a resource. Set the interrupted state and then wait for the run lock.
func (i *Queue) Queue() Unlocker {
	i.interrupt()
	i.runLock.Lock()

	i.running()
	return func() {
		i.done()
	}
}

func (i *Queue) running() {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	i.waiting--
}

func (i *Queue) interrupt() {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	i.waiting++
	log.Debug("Added to queue: ", i.waiting)
}

func (i *Queue) IsInterrupted() bool {
	i.interruptLock.Lock()
	defer i.interruptLock.Unlock()

	return i.waiting != 0
}

func (i *Queue) done() {
	defer i.runLock.Unlock()

	log.Debug("Marked done. Currently waiting: ", i.waiting)
	if i.waiting < 0 {
		log.Warn(errors.New("number waiting in queue less than zero"))
	}
}
