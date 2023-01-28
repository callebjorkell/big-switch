package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	Token   string
	BaseUrl *url.URL
	Caller  string
}

func NewClient(baseUrl, token, caller string) *Client {
	u, _ := url.Parse(baseUrl)

	return &Client{
		Token:   token,
		BaseUrl: u,
		Caller:  caller,
	}
}

func (c *Client) Do(r *http.Request, responseBody any) error {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.Token))
	r.Header.Set("X-Caller-Email", c.Caller)
	r.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if responseBody == nil {
		return nil
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, responseBody)
}

func (c *Client) NewStatusRequest(service, namespace string) (*http.Request, error) {
	values := url.Values{}
	values.Add("service", service)
	if namespace != "" {
		values.Add("namespace", namespace)
	}
	u := c.BaseUrl.JoinPath("status")
	u.RawQuery = values.Encode()

	return http.NewRequest("GET", u.String(), nil)
}

func (c *Client) NewPromoteRequest(service, artifactID string) (*http.Request, error) {
	type ReleaseRequest struct {
		Service        string          `json:"service"`
		Environment    string          `json:"environment"`
		ArtifactID     string          `json:"artifactId"`
		CommitterName  string          `json:"committerName"`
		CommitterEmail string          `json:"committerEmail"`
		Intent         json.RawMessage `json:"intent"`
	}

	releaseReq := ReleaseRequest{
		Service:        service,
		Environment:    "prod",
		ArtifactID:     artifactID,
		CommitterName:  "Surveyor deployer",
		CommitterEmail: c.Caller,
		Intent:         json.RawMessage(`{"type":"Promote","promote":{"fromEnvironment":"dev"}}`),
	}

	body, err := json.Marshal(releaseReq)
	if err != nil {
		return nil, fmt.Errorf("could not encode JSON body: %w", err)
	}
	bodyReader := bytes.NewReader(body)

	return http.NewRequest("POST", c.BaseUrl.JoinPath("release").String(), bodyReader)
}
