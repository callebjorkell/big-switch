//go:build !pi

package neopixel

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type mockEngine struct {
	colors []uint32
}

func (d mockEngine) Init() error {
	return nil
}

func (d mockEngine) Render() error {
	fmt.Println("neopixel: Render")
	log.Debugf("colors: %#v", d.colors)
	return nil
}

func (d mockEngine) Wait() error {
	fmt.Println("neopixel: Wait")
	return nil
}

func (d mockEngine) Fini() {
	fmt.Println("neopixel: Fini")
}

func (d mockEngine) Leds(_ int) []uint32 {
	return d.colors
}

func NewLedController() *LedController {
	return &LedController{
		ws: mockEngine{
			colors: make([]uint32, 1),
		},
	}
}
