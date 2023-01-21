//go:build pi

package lcd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
	"time"
)

func init() {
	if _, err := host.Init(); err != nil {
		logrus.Fatalln("Unable to initialize periph:", err)
	}
}

// InitLCD initializes all the LCD pins
func InitLCD() {
	logrus.Infoln("Initializing LCD")
	registerSelection = gpioreg.ByName(registerSelectionPin)
	clockEdge = gpioreg.ByName(clockEdgePin)
	dataPins[0] = gpioreg.ByName(data4Pin)
	dataPins[1] = gpioreg.ByName(data5Pin)
	dataPins[2] = gpioreg.ByName(data6Pin)
	dataPins[3] = gpioreg.ByName(data7Pin)

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
	for i, pin := range dataPins {
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

func Println(l Line, msg string) {
	sendByte(byte(l), command)
	m := fmt.Sprintf("%-16s", msg)
	for i := 0; i < lineWidth; i++ {
		sendByte(m[i], character)
	}
}

func Clear(l Line) {
	Println(l, "")
}
