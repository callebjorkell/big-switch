package deploy

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const MinRatePercentage = 20

var ErrRateLimited = errors.New("rate limiting safety margin has been hit")

type ChangeEvent struct {
	Service  string
	Artifact string
}

type Checker struct {
	Token      string
	killSwitch chan struct{}
	changes    chan ChangeEvent
	stopper    sync.Once
}

func (c *Checker) Changes() <-chan ChangeEvent {
	return c.changes
}

func (c *Checker) Close() error {
	c.stopper.Do(func() {
		close(c.killSwitch)
		close(c.changes)
	})
	return nil
}

func NewChecker(token string) *Checker {
	c := Checker{Token: token}
	log.Debug("Initializing the checker...")
	c.changes = make(chan ChangeEvent, 10)
	c.killSwitch = make(chan struct{})

	return &c
}

func (c *Checker) AddWatch(service string) error {
	go func() {
		log.Infof("Starting to watch %s", service)
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				// fall out of the select and do the work.
			case _, open := <-c.killSwitch:
				if !open {
					log.Debugf("Kill switch flipped. Stopping watch of %s", service)
					return
				}
			}

			a, err := c.GetArtifacts(service)
			if err != nil {
				log.Warnf("error when watching %s: %v", service, err)
				continue
			}

			if a.Dev.Name != a.Prod.Name {
				log.Debugf("%s prod (%s) differs from dev (%s)", service, a.Prod.Name, a.Dev.Name)

				if a.Prod.Time.After(a.Dev.Time) {
					log.Debugf("%s prod is newer than dev (%v later than %v). Not offering deploy.", service, a.Prod.Time, a.Dev.Time)
				}
				c.changes <- ChangeEvent{
					Service:  a.Service,
					Artifact: a.Prod.Name,
				}
			}
		}
	}()

	return nil
}

type Artifacts struct {
	Service string
	Prod    Artifact
	Dev     Artifact
}
type Artifact struct {
	Time time.Time
	Name string
}

func (c *Checker) GetArtifacts(service string) (Artifacts, error) {
	const timeLayout = "2006-01-02 15:04:05"

	// TODO: Get stuff from release manager.
	timestamp, _ := time.Parse(timeLayout, "2023-01-19 15:10:49")

	return Artifacts{
		Service: service,
		Dev: Artifact{
			Time: timestamp,
			Name: "master-f9235524df-a7faf42078",
		},
		Prod: Artifact{
			Time: timestamp,
			Name: "master-f9235524df-a7faf42078",
		},
	}, nil
}
