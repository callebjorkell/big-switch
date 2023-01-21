package deploy

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	setDebug()

	c := Checker{Token: "tok"}
	err := c.AddWatch("fraud-screening")

	<-time.After(60 * time.Second)
	e := <-c.changes
	log.Printf("Event: %v", e)
	c.Close()
	<-time.After(1 * time.Second)
	assert.NoError(t, err)
}

func setDebug() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	log.SetLevel(log.DebugLevel)
}
