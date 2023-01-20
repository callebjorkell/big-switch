package button

import "fmt"

type ButtonEvent struct {
	Pressed bool
}

func (b ButtonEvent) String() string {
	action := "pressed"
	if !b.Pressed {
		action = "released"
	}
	return fmt.Sprintf("Button was %v", action)
}
