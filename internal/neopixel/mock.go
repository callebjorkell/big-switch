//go:build !pi

package neopixel

type dummyEngine struct {
}

func (d dummyEngine) Init() error {
	panic("implement me")
}

func (d dummyEngine) Render() error {
	panic("implement me")
}

func (d dummyEngine) Wait() error {
	panic("implement me")
}

func (d dummyEngine) Fini() {
	panic("implement me")
}

func (d dummyEngine) Leds(channel int) []uint32 {
	panic("implement me")
}

func NewLedController() *LedController {
	return &LedController{
		ws: dummyEngine{},
	}
}
