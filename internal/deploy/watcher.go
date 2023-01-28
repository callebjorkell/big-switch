package deploy

import (
	"context"
	"github.com/callebjorkell/big-switch/internal/lcd"
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
		lastProdArtifact := Artifact{}

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

			if a.IsProdBehind() {
				if lastProdArtifact.Equals(a.Prod) {
					log.Debugf("Have already seen current prod artifact for %s. Skipping.", service)
					continue
				}
				
				log.Infof("%s prod (%s) differs from dev (%s)", service, a.Prod.Name, a.Dev.Name)
				lastProdArtifact = a.Prod

				w.changes <- ChangeEvent{
					Service:  a.Service,
					Artifact: a.Prod.Name,
				}
			}
		}
	}()

	return nil
}

func (w *Watcher) GetArtifacts(service, namespace string) (Artifacts, error) {
	type statusPayload struct {
		Environments []struct {
			Artifact
			Environment string `json:"name"`
		} `json:"environments"`
	}

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

type Notifier interface {
	Alert(service string)
	Success()
	Failure()
	Reset()
}

type Deployer interface {
	Promote(service, artifact string) error
}

func ChangeListener(
	ctx context.Context,
	notifier Notifier,
	promoter Deployer,
	changes <-chan ChangeEvent,
	confirm <-chan bool,
) {
	for {
		select {
		case <-ctx.Done():
			break
		case e := <-changes:
			log.Infof("Service %s changed. Waiting for confirmation!", e.Service)
			notifier.Alert(e.Service)

			select {
			case confirmed := <-confirm:
				if confirmed {
					log.Infof("Promoting %s for service %s to production.", e.Artifact, e.Service)
					err := promoter.Promote(e.Service, e.Artifact)
					if err != nil {
						log.Warn("Unable to trigger deploy: ", err)
						lcd.Print("TRIGGER FAILED", "")
						notifier.Failure()
						<-time.After(5 * time.Second)
					} else {
						notifier.Success()
					}
				}
			case <-time.After(45 * time.Second):
				log.Info("Confirmation timed out.")
			}
			notifier.Reset()
		}
	}
}
