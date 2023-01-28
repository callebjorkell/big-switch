package deploy

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var statusTemplate = `{
  "defaultNamespaces": true,
  "environments": [
    {
      "name": "dev",
      "tag": "{{ .DevArtifact }}",
      "committer": "GitHub",
      "author": "Awesome dev 1",
      "message": "this-is-my-test-pr-1337",
      "date": {{ .DevTime }},
      "buildUrl": "https://build.localhost/1337"
    },
    {
      "name": "prod",
      "tag": "{{ .ProdArtifact }}",
      "committer": "GitHub",
      "author": "Awesome dev 2",
      "message": "this-is-my-first-test-pr-1336",
      "date": {{ .ProdTime }},
      "buildUrl": "https://build.localhost/1336"
    }
  ]
}
`

type statusData struct {
	DevArtifact  string
	DevTime      int64
	ProdArtifact string
	ProdTime     int64
}

func TestWatch(t *testing.T) {
	setDebug()

	tmpl, err := template.New("status").Parse(statusTemplate)
	require.NoError(t, err)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		tmpl.Execute(w, statusData{
			DevArtifact:  "master-04169c5a19-a9c84eb8ff",
			DevTime:      1674119917135,
			ProdArtifact: "master-6831b4ba23-5876ec33b0",
			ProdTime:     1674119916510,
		})
	}
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	c := NewClient(s.URL, "arst", "me@local.com")
	w := NewWatcher(c)
	defer w.Close()
	err = w.AddWatch("some-service", "prod", time.Millisecond, 5*time.Millisecond)
	assert.NoError(t, err)

	select {
	case e := <-w.Changes():
		assert.Equal(t, "some-service", e.Service)
		assert.Equal(t, "master-6831b4ba23-5876ec33b0", e.Artifact)
	case <-time.After(250 * time.Millisecond):
		t.Fatal("timed out waiting for change event")
	}
}

func TestWatch_OnlyReportsChangeOnce(t *testing.T) {
	setDebug()

	tmpl, err := template.New("status").Parse(statusTemplate)
	require.NoError(t, err)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		tmpl.Execute(w, statusData{
			DevArtifact:  "master-04169c5a19-a9c84eb8ff",
			DevTime:      1674119917135,
			ProdArtifact: "master-6831b4ba23-5876ec33b0",
			ProdTime:     1674119916510,
		})
	}
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	c := NewClient(s.URL, "arst", "me@local.com")
	w := NewWatcher(c)
	defer w.Close()
	err = w.AddWatch("some-service", "prod", time.Millisecond, 5*time.Millisecond)
	assert.NoError(t, err)

	select {
	case <-w.Changes():
		select {
		case <-w.Changes():
			t.Fatal("only a single change should have been produced")
		case <-time.After(250 * time.Millisecond):
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("timed out waiting for change event")
	}
}

func TestReleaseRequestBody(t *testing.T) {
	c := NewClient("localhost", "arst", "me@local.com")
	req, err := c.NewPromoteRequest("test-service", "the-dev-artifact")

	require.NoError(t, err)
	require.NotNil(t, req)
	body, _ := io.ReadAll(req.Body)
	require.Equal(t, `{"service":"test-service","environment":"prod","artifactId":"the-dev-artifact","committerName":"Surveyor deployer","committerEmail":"me@local.com","intent":{"type":"Promote","promote":{"fromEnvironment":"dev"}}}`, string(body))
}

func setDebug() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	log.SetLevel(log.DebugLevel)
}
