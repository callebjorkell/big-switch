package lcd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
	"time"
)

type Line byte

func init() {
	if _, err := host.Init(); err != nil {
		logrus.Fatalln("Unable to initialize periph:", err)
	}
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
	datapins          [4]gpio.PinIO
)

// InitLCD initializes all the LCD pins
func InitLCD() {
	logrus.Infoln("Initializing LCD")
	registerSelection = gpioreg.ByName(registerSelectionPin)
	clockEdge = gpioreg.ByName(clockEdgePin)
	datapins[0] = gpioreg.ByName(data4Pin)
	datapins[1] = gpioreg.ByName(data5Pin)
	datapins[2] = gpioreg.ByName(data6Pin)
	datapins[3] = gpioreg.ByName(data7Pin)

	sendByte(0x33, command)
	sendByte(0x32, command)
	sendByte(0x28, command)
	sendByte(0x0C, command)
	sendByte(0x06, command)
	sendByte(0x01, command)
}

func sendByte(bits byte, mode gpio.Level) {
	registerSelection.Out(mode)
	pulseByte(bits, 0x10)
	pulseByte(bits, 0x01)
}

func pulseByte(bits, mask byte) {
	for i, pin := range datapins {
		pin.Out(gpio.Low)
		if bits&(mask<<uint(i)) != 0 {
			pin.Out(gpio.High)
		}
	}
	time.Sleep(signalDelay)
	clockEdge.Out(gpio.High)
	time.Sleep(signalPulse)
	clockEdge.Out(gpio.Low)
	time.Sleep(signalDelay)
}

func PrintLine(l Line, msg string) {
	sendByte(byte(l), command)
	m := fmt.Sprintf("%-16s", msg)
	for i := 0; i<lineWidth; i++ {
		sendByte(m[i], character)
	}
}

func Clear(l Line) {
	PrintLine(l, "")
}

func Reset() {
	PrintLine(Line1, "Awesome Deployer")
	Clear(Line2)
}

func ClearAll() {
	Clear(Line1)
	Clear(Line2)
}