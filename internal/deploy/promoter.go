package deploy

type Promoter struct {
	client *Client
}

func NewPromoter(c *Client) *Promoter {
	return &Promoter{
		client: c,
	}
}

func (p *Promoter) Promote(service, artifactID string) error {
	req, err := p.client.NewPromoteRequest(service, artifactID)
	if err != nil {
		return err
	}

	return p.client.Do(req, nil)
}
