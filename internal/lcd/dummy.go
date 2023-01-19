//go:build !pi

package lcd

import (
	"fmt"
)

// InitLCD initializes all the LCD pins
func InitLCD() {
	fmt.Println("Starting the LCD")
}

func PrintLine(l Line, msg string) {
	fmt.Printf(`Print line %v: "%v"\n`, l, msg)
}

func Clear(l Line) {
	fmt.Printf(`Clear line %v\n`, l)
}

func Reset() {
	PrintLine(Line1, "Awesome Deployer")
	Clear(Line2)
}

func ClearAll() {
	Clear(Line1)
	Clear(Line2)
}
