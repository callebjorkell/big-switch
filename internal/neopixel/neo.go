package neopixel

import (
	"fmt"
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
	ws      wsEngine
	stopper sync.Once
	queue   Queue
}

func (l *LedController) Stop() {
	log.Info("Stop animation.")
	done := l.queue.Queue()
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

func (f *LedController) setColor(color uint32) error {
	for i := 0; i < len(f.ws.Leds(0)); i++ {
		f.ws.Leds(0)[i] = color
	}
	if err := f.ws.Render(); err != nil {
		return err
	}
	return nil
}

func (l *LedController) Rainbow() error {
	done := l.queue.Queue()
	defer done()
	defer l.clear()

	log.Debugf("Displaying rainbow")
	tick := time.NewTicker(30 * time.Millisecond)
	defer tick.Stop()

	for step := 0; step < 400; step++ {
		if l.queue.IsInterrupted() {
			return fmt.Errorf("animtion was interrupted")
		}

		c := getRGB(step)
		if step > 300 {
			c = withBrightness(c, uint32(100-(step%100)))
		}

		err := l.setColor(c)
		if err != nil {
			return err
		}

		<-tick.C
	}
	
	return nil
}

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

func (l *LedController) Flash(color uint32) {
	done := l.queue.Queue()
	defer done()

	log.Infof("Flashing color %06x", color)

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
	done := l.queue.Queue()

	go func() {
		defer done()
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
	log.Debugf("Breathing color: %06x", color)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if l.queue.IsInterrupted() {
			log.Debug("Animation interrupted.")
			return fmt.Errorf("animtion is stopped")
		}

		c := withBrightness(color, light)

		err := l.setColor(c)
		if err != nil {
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

func toRGB(r, g, b uint32) uint32 {
	return (r << 16) | (g << 8) | b
}
