package deploy

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommit(t *testing.T) {
	c := Checker{Token: "tok"}
	c.GetHeadCommit("Tradeshift", "truebn-sparkles", "master")
}

func TestWatch(t *testing.T) {
	setDebug()

	c := Checker{Token: "tok"}
	err := c.AddWatch("Tradeshift", "truebn-sparkles", "master")
	err2 := c.AddWatch("callebjorkell", "big-switch", "test")

	<-time.After(60 * time.Second)
	e := <-c.changes
	log.Printf("Event: %v", e)
	c.Close()
	<-time.After(1 * time.Second)
	assert.Nil(t, err)
	assert.Nil(t, err2)
}

func setDebug() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	log.SetLevel(log.DebugLevel)
}
