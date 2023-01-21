//go:build !pi

package lcd

import (
	"fmt"
)

func InitLCD() {
	fmt.Println("Starting the LCD")
}

func Println(l Line, msg string) {
	fmt.Printf("Print line %v: \"%v\"\n", l, msg)
}

func Clear(l Line) {
	fmt.Printf("Clear line %v\n", l)
}
