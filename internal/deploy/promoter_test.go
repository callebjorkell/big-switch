package deploy

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var alreadyUpToDate = `{
  "service": "test-service",
  "status": "Environment 'prod' is already up-to-date",
  "toEnvironment":"prod"
}`

func TestPromote(t *testing.T) {
	setDebug()

	handler := func(w http.ResponseWriter, r *http.Request) {
		bodyStr, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		body := map[string]any{}
		json.Unmarshal(bodyStr, &body)

		assert.Equal(t, "test-service", body["service"])
		assert.Equal(t, "some-artifact", body["artifactId"])
		assert.NotNil(t, body["intent"])
		assert.NotEmpty(t, body["committerName"])
		assert.NotNil(t, body["committerEmail"])

		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(alreadyUpToDate))
	}
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	c := NewClient(s.URL, "asdf", "me@local.com")
	p := NewPromoter(c)

	err := p.Promote("test-service", "some-artifact")
	assert.NoError(t, err)
}

func TestChangeListenerNotArmed(t *testing.T) {
	confirmations := make(chan bool)
	changes := make(chan ChangeEvent)
	notifier := NewNotifierMock()
	promoter := NewPromoterMock(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ChangeListener(ctx, notifier, promoter, changes, confirmations)

	select {
	case confirmations <- true:
	default:
	}

	assert.True(t, promoter.NoInteraction())
}

func TestChangeListenerAlerting(t *testing.T) {
	confirmations := make(chan bool)
	changes := make(chan ChangeEvent)
	notifier := NewNotifierMock()
	promoter := NewPromoterMock(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ChangeListener(ctx, notifier, promoter, changes, confirmations)

	changes <- ChangeEvent{Service: "test-service", Artifact: "some-artifact"}

	notifier.WaitForInteraction(50 * time.Millisecond)

	assert.True(t, promoter.NoInteraction())
	assert.Equal(t, "test-service", notifier.alertFor)
}

func TestChangeListenerDeploying(t *testing.T) {
	confirmations := make(chan bool)
	changes := make(chan ChangeEvent)
	notifier := NewNotifierMock()
	promoter := NewPromoterMock(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ChangeListener(ctx, notifier, promoter, changes, confirmations)

	changes <- ChangeEvent{Service: "test-service", Artifact: "some-artifact"}
	notifier.WaitForInteraction(50 * time.Millisecond)

	confirmations <- true
	promoter.WaitForInteraction(80 * time.Millisecond)

	assert.Equal(t, "test-service", promoter.service)
	assert.Equal(t, "some-artifact", promoter.artifact)
	assert.Equal(t, "test-service", notifier.alertFor)
}

type NotifierMock struct {
	alertFor                string
	interactionChan         chan bool
	success, failure, reset bool
}

func NewNotifierMock() *NotifierMock {
	return &NotifierMock{
		interactionChan: make(chan bool),
	}
}

func (n *NotifierMock) WaitForInteraction(timeout time.Duration) bool {
	select {
	case <-n.interactionChan:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (n *NotifierMock) Alert(service string) {
	n.alertFor = service
	n.interactionChan <- true
}

func (n *NotifierMock) Success() {
	n.success = true
	n.interactionChan <- true
}

func (n *NotifierMock) Failure() {
	n.failure = true
	n.interactionChan <- true
}

func (n *NotifierMock) Reset() {
	n.reset = true
	n.interactionChan <- true
}

type PromoterMock struct {
	retErr            error
	service, artifact string
	interactionChan   chan bool
}

func NewPromoterMock(returnedError error) *PromoterMock {
	return &PromoterMock{
		retErr:          returnedError,
		interactionChan: make(chan bool),
	}
}

func (p *PromoterMock) Promote(service, artifact string) error {
	p.service = service
	p.artifact = artifact

	p.interactionChan <- true
	return p.retErr
}

func (p *PromoterMock) NoInteraction() bool {
	select {
	case <-p.interactionChan:
	default:
	}

	return p.service == "" && p.artifact == ""
}

func (p *PromoterMock) WaitForInteraction(timeout time.Duration) bool {
	select {
	case <-p.interactionChan:
		return true
	case <-time.After(timeout):
		return false
	}
}
