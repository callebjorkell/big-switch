//go:build pi

package lcd

import (
	"fmt"
	"testing"
)

func TestRange(t *testing.T) {
	bits := 0x30
	for i := range datapins {
		fmt.Println(i, " to low")
		fmt.Printf("0x%x\n", 0x10<<uint(i))
		if bits&(0x10<<uint(i)) != 0 {
			fmt.Println(i, " to high")
		}
	}
}
