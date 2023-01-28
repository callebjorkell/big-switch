package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Start a test HTTP server that can be used as a mock for the release manager

var statusTemplate = `{
  "defaultNamespaces": true,
  "environments": [
    {
      "name": "dev",
      "tag": "{{ .DevArtifact }}",
      "committer": "GitHub",
      "date": {{ .DevTime.UnixMilli }}
    },
    {
      "name": "prod",
      "tag": "{{ .ProdArtifact }}",
      "committer": "GitHub",
      "date": {{ .ProdTime.UnixMilli }}
    }
  ]
}
`

var alreadyUpToTemplate = `{
  "service": "{{ .Service }}",
  "status": "Environment 'prod' is already up-to-date",
  "toEnvironment":"prod"
}`

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	handlers := NewHandlers()

	log.Info("Starting release-manager test server at :9090")
	server := http.Server{Addr: ":9090"}
	http.HandleFunc("/status", handlers.statusHandler)
	http.HandleFunc("/release", handlers.promoteHandler)
	log.Infof("Starting test-server on %v.", server.Addr)

	go func() {
		<-signalChan
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	server.ListenAndServe()
}

type Handlers struct {
	DevArtifact  string
	DevTime      time.Time
	ProdArtifact string
	ProdTime     time.Time
	Service      string
	lock         sync.Mutex
}

func NewHandlers() *Handlers {
	now := time.Now()

	return &Handlers{
		DevArtifact:  fmt.Sprintf("artifact-%d", now.Unix()),
		DevTime:      now,
		ProdArtifact: fmt.Sprintf("artifact-%d", now.Unix()),
		ProdTime:     now,
		Service:      "little-test-service",
		lock:         sync.Mutex{},
	}
}

func (h *Handlers) promoteHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqDump, _ := httputil.DumpRequest(req, true)
	log.Debugf("serving promote request: %s", string(reqDump))

	w.Header().Set("Content-Type", "application/json")
	tmpl, err := template.New("promote").Parse(alreadyUpToTemplate)
	if err != nil {
		w.WriteHeader(500)
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	h.ProdTime = h.DevTime
	h.ProdArtifact = h.DevArtifact

	tmpl.Execute(w, h)
}

func (h *Handlers) statusHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqDump, _ := httputil.DumpRequest(req, true)
	log.Debugf("serving status request: %s", string(reqDump))

	w.Header().Set("Content-Type", "application/json")
	tmpl, err := template.New("status").Parse(statusTemplate)
	if err != nil {
		w.WriteHeader(500)
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	if h.DevTime.Add(2 * time.Minute).Before(time.Now()) {
		h.DevTime = time.Now()
		h.DevArtifact = fmt.Sprintf("artifact-%d", h.DevTime.Unix())
	}

	tmpl.Execute(w, h)
}
