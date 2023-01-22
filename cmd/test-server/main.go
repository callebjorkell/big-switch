package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
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
      "date": {{ .DevTime }}
    },
    {
      "name": "prod",
      "tag": "{{ .ProdArtifact }}",
      "committer": "GitHub",
      "date": {{ .ProdTime }}
    }
  ]
}
`

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Starting release-manager test server at :9090")
	server := http.Server{Addr: ":9090"}
	http.HandleFunc("/status", statusHandler())
	log.Infof("Starting server on %v. Waiting for passphrase.", server.Addr)

	go func() {
		<-signalChan
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	server.ListenAndServe()
}

type statusData struct {
	DevArtifact  string
	DevTime      int64
	ProdArtifact string
	ProdTime     int64
}

func statusHandler() func(w http.ResponseWriter, req *http.Request) {
	prodTime := time.Now()
	devTime := prodTime
	l := sync.Mutex{}

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tmpl, err := template.New("status").Parse(statusTemplate)
		if err != nil {
			w.WriteHeader(500)
		}

		l.Lock()
		defer l.Unlock()
		if prodTime.Add(3 * time.Minute).Before(time.Now()) {
			prodTime = devTime
		}
		if devTime.Add(2 * time.Minute).Before(time.Now()) {
			devTime = time.Now()
		}
		s := statusData{
			ProdArtifact: "master-04169c5a19-a9c84eb8ff",
			ProdTime:     prodTime.UnixMilli(),
			DevTime:      devTime.UnixMilli(),
		}
		if prodTime.Equal(devTime) {
			s.DevArtifact = s.ProdArtifact
		} else {
			s.DevArtifact = "master-6831b4ba23-5876ec33b0"
		}

		tmpl.Execute(w, s)
	}
}
