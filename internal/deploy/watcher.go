package deploy

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

var pollingInterval = 30 * time.Second

type ChangeEvent struct {
	Service  string
	Artifact string
}

type Watcher struct {
	Caller     string
	killSwitch func()
	ctx        context.Context
	changes    chan ChangeEvent
	client     *Client
}

func (w *Watcher) Changes() <-chan ChangeEvent {
	return w.changes
}

func (w *Watcher) Close() error {
	w.killSwitch()
	return nil
}

func NewWatcher(client *Client) *Watcher {
	log.Debug("Initializing the checker...")
	ctx, cancel := context.WithCancel(context.Background())
	c := Watcher{
		ctx:        ctx,
		killSwitch: cancel,
		client:     client,
	}
	c.changes = make(chan ChangeEvent, 10)

	return &c
}

func (w *Watcher) AddWatch(service, namespace string) error {
	go func() {
		log.Infof("Starting to watch %s", service)
		t := time.NewTicker(pollingInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				// fall out of the select and do the work.
			case <-w.ctx.Done():
				log.Infof("Stopping watch of %s", service)
				close(w.changes)
				return
			}

			a, err := w.GetArtifacts(service, namespace)
			if err != nil {
				log.Warnf("error when watching %s: %v", service, err)
				continue
			}

			if a.Dev.Name != a.Prod.Name {
				log.Debugf("%s prod (%s) differs from dev (%s)", service, a.Prod.Name, a.Dev.Name)

				if a.Prod.Time > a.Dev.Time {
					log.Debugf("%s prod is newer than dev (%v later than %v). Not offering deploy.", service, a.Prod.Time, a.Dev.Time)
				}
				w.changes <- ChangeEvent{
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
	Time int64  `json:"date"`
	Name string `json:"tag"`
}

type statusPayload struct {
	Environments []struct {
		Artifact
		Environment string `json:"name"`
	} `json:"environments"`
}

func (w *Watcher) GetArtifacts(service, namespace string) (Artifacts, error) {
	req, err := w.client.NewStatusRequest(service, namespace)
	if err != nil {
		return Artifacts{}, err
	}

	status := statusPayload{}
	w.client.Do(req, &status)
	if err != nil {
		return Artifacts{}, err
	}

	a := Artifacts{
		Service: service,
	}

	for _, env := range status.Environments {
		switch env.Environment {
		case "dev":
			a.Dev = env.Artifact
		case "prod":
			a.Prod = env.Artifact
		}
	}

	log.Debug(a)
	return a, nil
}
