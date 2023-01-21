package button

import "fmt"

type Event struct {
	Pressed bool
}

func (b Event) String() string {
	action := "pressed"
	if !b.Pressed {
		action = "released"
	}
	return fmt.Sprintf("Button was %v", action)
}
