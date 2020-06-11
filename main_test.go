package main

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	c, err := getConfig()
	assert.Nil(t, err)
	assert.Equal(t, "arst", c.Github.Token)
	assert.Equal(t, "tsra", c.Jenkins.Token)
	assert.Equal(t, "callebjorkell", c.Repositories[1].Owner)

	log.Print(c)
}
