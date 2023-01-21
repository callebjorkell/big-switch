package lcd

import (
	"fmt"
	"periph.io/x/conn/v3/gpio"
	"strings"
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

// Center aligns a string to the 16 character window size. If the string is longer than 16 characters, it will be
// truncated to fit.
func Center(msg string) string {
	if len(msg) >= 16 {
		return msg[:16]
	}
	leftPad := (16 - len(msg)) / 2
	return fmt.Sprintf("%v%v", strings.Repeat(" ", leftPad), msg)
}

func Reset() {
	Println(Line1, "Awesome Deployer")
	Clear(Line2)
}

func ClearAll() {
	Clear(Line1)
	Clear(Line2)
}
