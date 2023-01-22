package deploy

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var pollingInterval = 30 * time.Second

type ChangeEvent struct {
	Service  string
	Artifact string
}

type Watcher struct {
	Token      string
	BaseUrl    *url.URL
	Caller     string
	killSwitch chan struct{}
	changes    chan ChangeEvent
	stopper    sync.Once
}

func (w *Watcher) Changes() <-chan ChangeEvent {
	return w.changes
}

func (w *Watcher) Close() error {
	w.stopper.Do(func() {
		close(w.killSwitch)
	})
	return nil
}

func NewWatcher(baseUrl, token, caller string) *Watcher {
	log.Debug("Initializing the checker...")
	u, _ := url.Parse(baseUrl)
	c := Watcher{
		Token:   token,
		BaseUrl: u,
		Caller:  caller,
	}
	c.changes = make(chan ChangeEvent, 10)
	c.killSwitch = make(chan struct{})
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
			case <-w.killSwitch:
				log.Infof("Kill switch flipped. Stopping watch of %s", service)
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
	const timeLayout = "2006-01-02 15:04:05"

	values := url.Values{}
	values.Add("service", service)
	if namespace != "" {
		values.Add("namespace", namespace)
	}
	u := w.BaseUrl.JoinPath("status")
	u.RawQuery = values.Encode()
	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return Artifacts{}, err
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", w.Token))
	r.Header.Set("X-Caller-Email", w.Caller)
	r.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return Artifacts{}, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return Artifacts{}, err
	}
	status := statusPayload{}
	err = json.Unmarshal(payload, &status)
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
