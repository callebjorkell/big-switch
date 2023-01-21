package neopixel

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

func (l *LedController) Flash(color uint32) {
	done := l.interruptor.Interrupt()
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

func (l *LedController) Rainbow() error {
	done := l.interruptor.Interrupt()
	defer done()
	defer l.clear()

	log.Debugf("Displaying rainbow")
	tick := time.NewTicker(30 * time.Millisecond)
	defer tick.Stop()

	for step := 0; step <= 450; step++ {
		if l.interruptor.IsInterrupted() {
			return fmt.Errorf("animtion was interrupted")
		}

		c := getRGB(step)
		if step < 50 {
			c = withBrightness(c, uint32(step*2))
		}
		if step > 350 {
			c = withBrightness(c, uint32(450-step))
		}

		err := l.setColor(c)
		if err != nil {
			return err
		}

		<-tick.C
	}

	return nil
}

func (l *LedController) Breathe(color uint32) {
	done := l.interruptor.Interrupt()

	go func() {
		defer done()
		defer l.clear()
		for {
			err := l.singleBreath(color)
			if err != nil {
				log.Debug("Stopping breathing: ", err)
				break
			}
		}
	}()
}

func (l *LedController) singleBreath(color uint32) error {
	light := uint32(0)
	increase := true
	log.Debugf("Breathing color: %06x", color)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if l.interruptor.IsInterrupted() {
			log.Debug("Animation interrupted.")
			return fmt.Errorf("animtion is interrupted")
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
