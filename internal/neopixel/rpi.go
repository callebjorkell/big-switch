//go:build pi

package neopixel

import (
	ws "github.com/rpi-ws281x/rpi-ws281x-go"
)

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
		ws: dev,
	}
}
