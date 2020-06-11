package neopixel

import (
	"errors"
	ws "github.com/rpi-ws281x/rpi-ws281x-go"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	brightness = 250
	ledCounts  = 24
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

type LedController struct {
	ws     wsEngine
	once   *sync.Once
	stop   chan bool
	lock   *sync.Mutex
}

func NewLedController() *LedController {
	opt := ws.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCounts

	dev, err := ws.MakeWS2811(&opt)
	if err != nil {
		panic(err)
	}
	err = dev.Init()
	if err != nil {
		panic(err)
	}

	return &LedController{
		ws:     dev,
		stop:   make(chan bool),
		lock:   &sync.Mutex{},
		once:   &sync.Once{},
	}
}

func (l *LedController) Stop() {
	log.Info("Interrupt animation.")
	l.stop <- true
}

func (l *LedController) Close() error {
	l.once.Do(func() {
		log.Info("Stopping LED controller")
		close(l.stop)
		l.ws.Fini()
	})
	return nil
}

func (f *LedController) setColor(color uint32) error {
	for i := 0; i < len(f.ws.Leds(0)); i++ {
		f.ws.Leds(0)[i] = color
	}
	if err := f.ws.Render(); err != nil {
		return err
	}
	return nil
}

func (l *LedController) Flash(color uint32) {
	l.Stop()
	l.lock.Lock()
	log.Infof("Flashing color %06x", color)
	defer l.lock.Unlock()

	l.setColor(color)
	<-time.After(250 * time.Millisecond)
	l.setColor(0)
	<-time.After(40 * time.Millisecond)
	l.setColor(color)
	<-time.After(100 * time.Millisecond)
	l.setColor(0)
	<-time.After(40 * time.Millisecond)
	l.setColor(color)
	<-time.After(100 * time.Millisecond)
	l.setColor(0)

	log.Debug("Flashing done...")
}

func (l *LedController) clear() {
	l.setColor(0)
}

func (l *LedController) Breathe(color uint32) {
	l.Stop()
	l.lock.Lock()
	go func() {
		defer l.lock.Unlock()
		defer l.clear()
		for {
			err := l.singleBreathe(color)
			if err != nil {
				log.Debug("Stopping breathing: ", err)
				break
			}
		}
	}()
}

func (l *LedController) singleBreathe(color uint32) error {
	light := uint32(0)
	increase := true
	log.Infof("Breathing color: %06x", color)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-l.stop:
			log.Debug("Animation stopped.")
			return errors.New("animtion is stopped")
		default:
		}

		c := withBrightness(color, light)

		for i := 0; i < len(l.ws.Leds(0)); i++ {
			l.ws.Leds(0)[i] = c
		}
		if err := l.ws.Render(); err != nil {
			return err
		}

		if increase {
			light++
			if light > 100 {
				increase = false
			}
		} else {
			if light == 0 {
				break
			}
			light--
		}

		<-tick.C
	}
	return nil
}

// Get the same color, but with a lower or equal brightness, on a scale from 0-100, where 100 is the same as the input.
func withBrightness(color, light uint32) uint32 {
	if light >= 100 {
		return color
	}
	if light == 0 {
		return 0
	}

	r, g, b := (color>>16)&0xff, (color>>8)&0xff, color&0xff

	red := r * light / 100
	green := g * light / 100
	blue := b * light / 100

	return (red << 16) | (green << 8) | blue
}
