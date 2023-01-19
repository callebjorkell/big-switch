package neopixel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestColor(t *testing.T) {
	tt := []struct {
		name string
		input uint32
		light uint32
		output uint32
	}{
		{
			"full brightness red",
			0xff0000,
			100,
			0xff0000,
		},
		{
			"full brightness green",
			0x00ff00,
			100,
			0x00ff00,
		},
		{
			"full brightness blue",
			0x0000ff,
			100,
			0x0000ff,
		},
		{
			"zero brightness red",
			0xff0000,
			0,
			0x000000,
		},
		{
			"zero brightness green",
			0x00ff00,
			0,
			0x000000,
		},
		{
			"zero brightness blue",
			0x0000ff,
			0,
			0x000000,
		},
		{
			"50 percent",
			0x806040,
			50,
			0x403020,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			o := withBrightness(tc.input, tc.light)
			assert.Equal(t, tc.output, o)
		})
	}
}
