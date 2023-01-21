package neopixel

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

const (
	brightness = 250
	ledCounts  = 24
	ColorRed   = 0xFF0000
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

type LedController struct {
	ws          wsEngine
	stopper     sync.Once
	interruptor Interruptor
}

func (l *LedController) Stop() {
	log.Info("Stop animation.")
	done := l.interruptor.Interrupt()
	defer done()

	l.clear()
}

func (l *LedController) Close() error {
	l.stopper.Do(func() {
		log.Info("Stopping LED controller")
		l.Stop()
		l.ws.Fini()
	})
	return nil
}

func (l *LedController) setColor(color uint32) error {
	for i := 0; i < len(l.ws.Leds(0)); i++ {
		l.ws.Leds(0)[i] = color
	}
	if err := l.ws.Render(); err != nil {
		return err
	}
	return nil
}

// getRGB gets an RGB color, based on HSV, with angles aligned to 300 degrees for ease of calculation.
func getRGB(angle int) uint32 {
	a := uint32(angle % 300)
	if a <= 50 {
		return toRGB(255, a*5, 0)
	}
	if a <= 100 {
		return toRGB((100-a)*5, 255, 0)
	}
	if a <= 150 {
		return toRGB(0, 255, (a-100)*5)
	}
	if a <= 200 {
		return toRGB(0, (200-a)*5, 255)
	}
	if a <= 250 {
		return toRGB((a-200)*5, 0, 255)
	}
	return toRGB(255, 0, (300-a)*5)
}

func (l *LedController) clear() {
	l.setColor(0)
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

func toRGB(r, g, b uint32) uint32 {
	return (r << 16) | (g << 8) | b
}
