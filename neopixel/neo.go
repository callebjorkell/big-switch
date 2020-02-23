package neopixel

import (
	ws "github.com/rpi-ws281x/rpi-ws281x-go"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	brightness = 250
	ledCounts  = 24
	sleepTime  = 100
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type colorWipe struct {
	ws wsEngine
}

func (cw *colorWipe) setup() error {
	return cw.ws.Init()
}

func (cw *colorWipe) display(color uint32) error {
	for i := 0; i < len(cw.ws.Leds(0)); i++ {
		cw.ws.Leds(0)[i] = color
		if err := cw.ws.Render(); err != nil {
			return err
		}
		time.Sleep(sleepTime * time.Millisecond)
	}
	return nil
}

func Test() {
	opt := ws.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCounts

	dev, err := ws.MakeWS2811(&opt)
	checkError(err)

	//cw := &colorWipe{
	//	ws: dev,
	//}
	cw := &breathing{
		ws: dev,
	}
	checkError(cw.setup())
	defer dev.Fini()

	cw.display(uint32(0x0000ff))
	cw.display(uint32(0x00ff00))
	cw.display(uint32(0xff0000))
	cw.clear()
}

type breathing struct {
	ws wsEngine
}

func (b *breathing) setup() error {
	return b.ws.Init()
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

func (b *breathing) display(color uint32) error {
	light := uint32(0)
	increase := true
	log.Infof("Breathing color: %06x", color)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
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

	r,g,b := (color >> 16) & 0xff, (color >> 8) & 0xff, color & 0xff

	red := r * light / 100
	green := g * light / 100
	blue := b * light / 100

	return (red << 16) | (green << 8) | blue
}