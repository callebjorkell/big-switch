//go:build pi

package neopixel

import (
	ws "github.com/rpi-ws281x/rpi-ws281x-go"
)

func Test() {
	opt := ws.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCounts

	dev, err := ws.MakeWS2811(&opt)
	checkError(err)

	cw := &colorWipe{
		ws: dev,
	}
	checkError(cw.setup())
	defer dev.Fini()

	cw.display(uint32(0x0000ff))
	cw.display(uint32(0x00ff00))
	cw.display(uint32(0xff0000))
	cw.display(uint32(0x000000))
}
