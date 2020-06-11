package neopixel

import (
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

type flasher struct {
	ws    wsEngine
	color uint32
}

func (f *flasher) setColor(color uint32) error {
	for i := 0; i < len(f.ws.Leds(0)); i++ {
		f.ws.Leds(0)[i] = color
	}
	log.Debug("1")
	if err := f.ws.Render(); err != nil {
		return err
	}
	log.Debug("2")
	return nil
}

func (f *flasher) display() error {
	log.Debug("Display")
	f.setColor(f.color)
	<-time.After(700 * time.Millisecond)
	f.setColor(0)
	<-time.After(1000 * time.Millisecond)
	f.setColor(f.color)
	<-time.After(400 * time.Millisecond)
	f.setColor(0)
	<-time.After(500 * time.Millisecond)
	f.setColor(f.color)
	<-time.After(400 * time.Millisecond)
	f.setColor(0)
	return nil
}

func Flash(color uint32) {
	dev := initLeds()

	cw := &flasher{
		ws:    dev,
		color: color,
	}

	log.Infof("Flashing color %06x", color)
	cw.display()
	log.Debug("Flashing done...")
}

func initLeds() *ws.WS2811 {
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
	return dev
}

func NewCloser(c chan struct{}) *closer {
	return &closer{
		c,
		&sync.Once{},
	}
}

type closer struct {
	c    chan struct{}
	once *sync.Once
}

func (c closer) Close() error {
	log.Info("Stopping animation.")
	c.once.Do(func() {
		close(c.c)
	})
	return nil
}

func Breathe(color uint32) *closer {
	dev := initLeds()

	cw := &breathing{
		ws: dev,
	}

	c := make(chan struct{})
	go func() {
		defer dev.Fini()
		defer cw.clear()
		for {
			cw.display(color, c)
			select {
			case <-c:
				log.Debug("Done channel triggered.")
				return
			default:
			}
		}
	}()

	return NewCloser(c)
}

type breathing struct {
	ws wsEngine
}

func (b *breathing) clear() error {
	for i := 0; i < len(b.ws.Leds(0)); i++ {
		b.ws.Leds(0)[i] = 0
	}
	if err := b.ws.Render(); err != nil {
		return err
	}
	return nil
}

func (b *breathing) display(color uint32, stop <-chan struct{}) error {
	light := uint32(0)
	increase := true
	log.Infof("Breathing color: %06x", color)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-stop:
			log.Debug("Animation not active.")
			return nil
		default:
		}

		c := withBrightness(color, light)

		for i := 0; i < len(b.ws.Leds(0)); i++ {
			b.ws.Leds(0)[i] = c
		}
		if err := b.ws.Render(); err != nil {
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
