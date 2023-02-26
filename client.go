package nordigen

import (
	"context"
	"net/http"
	"net/url"

	"github.com/antonioplacerda/requests"
	uctx "github.com/antonioplacerda/requests/context"
)

// Option defines a function used to set options for the Client.
type Option func(c *Client)

// WithHTTPClient returns an Option that specifies the HTTP client to use
// as the basis of communications.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.hc = client
	}
}

func WithAuditor(auditor requests.Auditor) Option {
	return func(c *Client) {
		c.auditor = auditor
	}
}

// Client holds the integration with the Nordigen API.
type Client struct {
	hc      *http.Client
	baseURL *url.URL
	token   Token
	auditor requests.Auditor
}

// New returns a new Client.
func New(auditor requests.Auditor, options ...Option) (*Client, error) {
	baseURL, _ := url.Parse("https://ob.nordigen.com")
	client := &Client{
		hc:      http.DefaultClient,
		baseURL: baseURL,
		auditor: auditor,
	}

	for _, opt := range options {
		opt(client)
	}

	return client, nil
}

func (c *Client) request(ctx context.Context, token *Token, action string) *requests.Request {
	r := requests.NewRequest(c.baseURL, c.hc, "application/json").
		WithAuthorization(requests.NewBearerAuth(token.Access))
	if c.auditor != nil {
		r.WithAuditor(requests.NewReqAuditor(c.auditor, uctx.UserID(ctx), action))
	}
	return r
}
