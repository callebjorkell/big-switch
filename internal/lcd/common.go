package lcd

import (
	"periph.io/x/conn/v3/gpio"
	"time"
)

type Line byte

func (l Line) String() string {
	switch l {
	case Line1:
		return "L1"
	case Line2:
		return "L2"
	}
	return "N/A"
}

const (
	registerSelectionPin = "GPIO4"
	clockEdgePin         = "GPIO17"
	data4Pin             = "GPIO25"
	data5Pin             = "GPIO22"
	data6Pin             = "GPIO23"
	data7Pin             = "GPIO24"

	Line1 Line = 0x80
	Line2 Line = 0xC0

	lineWidth   = 16
	character   = gpio.High
	command     = gpio.Low
	signalPulse = 500000 * time.Nanosecond
	signalDelay = 500000 * time.Nanosecond
)

var (
	registerSelection gpio.PinIO
	clockEdge         gpio.PinIO
	dataPins          [4]gpio.PinIO
)
