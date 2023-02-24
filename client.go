package go_nordigen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Auditor audits the http request and corresponding response.
type Auditor interface {
	// ID returns the request ID.
	ID() string
	// Request audits the http Request.
	Request(id string, req *http.Request)
	// Response audits the http Response.
	Response(id string, resp *http.Response)
}

// Option defines a function used to set options for the Client.
type Option func(c *Client)

// WithHTTPClient returns an Option that specifies the HTTP client to use
// as the basis of communications.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.hc = client
	}
}

func WithAuditor(auditor Auditor) Option {
	return func(c *Client) {
		c.auditor = auditor
	}
}

func WithToken(token Token) Option {
	return func(c *Client) {
		c.token = token
	}
}

type BearerAuthorization struct {
	accessToken string
}

func (a *BearerAuthorization) Authorize(req *http.Request) (*http.Request, error) {
	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	return req, nil
}

// Client holds the integration with the Nordigen API.
type Client struct {
	hc      *http.Client
	baseURL *url.URL
	token   Token
	auditor Auditor
}

// New returns a new Client.
func New(options ...Option) (*Client, error) {
	baseURL, _ := url.Parse("https://ob.nordigen.com")
	client := &Client{
		hc:      http.DefaultClient,
		baseURL: baseURL,
		auditor: &reqAuditor{},
	}

	for _, opt := range options {
		opt(client)
	}

	return client, nil
}

type NewTokenRequest struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

type Token struct {
	Access         string `json:"access"`
	AccessExpires  int    `json:"access_expires"`
	Refresh        string `json:"refresh"`
	RefreshExpires int    `json:"refresh_expires"`
}

func (c *Client) NewToken(ctx context.Context, secretID, secretKey string) (*Token, error) {
	var res *Token
	_, err := NewRequest(c.baseURL, c.hc, c.auditor).
		WithJSONBody(NewTokenRequest{
			SecretID:  secretID,
			SecretKey: secretKey,
		}).
		WithResult(&res).
		Post(ctx, "/api/v2/token/new/")

	// {"secret_id":["This field may not be blank."],"secret_key":["This field may not be blank."],"status_code":400}
	return res, unwrapError(err)
}

type Institution struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Bic                  string   `json:"bic"`
	TransactionTotalDays string   `json:"transaction_total_days"`
	Countries            []string `json:"countries"`
	LogoURL              string   `json:"logo"`
}

func (c *Client) ListInstitutions(ctx context.Context, country string) ([]Institution, error) {
	var res []Institution
	_, err := NewRequest(c.baseURL, c.hc, c.auditor).
		WithAuthorization(&BearerAuthorization{c.token.Access}).
		WithQParam("country", country).
		WithResult(&res).
		Get(ctx, "/api/v2/institutions/")

	return res, unwrapError(err)
}

func (c *Client) CreateRequisition(ctx context.Context, redirectURL *url.URL, institutionID string, optional *CreateRequisitionOptional) (*Requisition, error) {
	var res *Requisition
	_, err := NewRequest(c.baseURL, c.hc, c.auditor).
		WithAuthorization(&BearerAuthorization{c.token.Access}).
		WithJSONBody(CreateRequisitionRequest{
			RedirectURL:               redirectURL.String(),
			InstitutionID:             institutionID,
			CreateRequisitionOptional: optional,
		}).
		WithResult(&res).
		Post(ctx, "/api/v2/requisitions/")

	return res, unwrapError(err)
}

type CreateRequisitionOptional struct {
	Agreement         string `json:"agreement,omitempty"`
	Reference         string `json:"reference,omitempty"`
	UserLanguage      string `json:"user_language,omitempty"`
	Ssn               string `json:"ssn,omitempty"`
	AccountSelection  bool   `json:"account_selection,omitempty"`
	RedirectImmediate bool   `json:"redirect_immediate,omitempty"`
}

type CreateRequisitionRequest struct {
	RedirectURL   string `json:"redirect"`
	InstitutionID string `json:"institution_id"`
	*CreateRequisitionOptional
}

type Requisition struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"-"`
	Status    string    `json:"status"`
	Accounts  []string  `json:"accounts"`
	Link      *url.URL  `json:"-"`
	CreateRequisitionRequest
}

func (c *Client) GetRequisition(ctx context.Context, requisitionID string) (*Requisition, error) {
	var res *Requisition
	_, err := NewRequest(c.baseURL, c.hc, c.auditor).
		WithAuthorization(&BearerAuthorization{c.token.Access}).
		WithResult(&res).
		Get(ctx, fmt.Sprintf("/api/v2/requisitions/%s", requisitionID))

	return res, unwrapError(err)
}

func (c *Client) GetBookedTransactions(ctx context.Context, accountID string) ([]Transaction, error) {
	var res *TransactionsResponse
	_, err := NewRequest(c.baseURL, c.hc, c.auditor).
		WithAuthorization(&BearerAuthorization{c.token.Access}).
		WithResult(&res).
		Get(ctx, fmt.Sprintf("/api/v2/accounts/%s/transactions/", accountID))

	return res.Transactions.Booked, unwrapError(err)
}

type TransactionsResponse struct {
	Transactions struct {
		Booked  []Transaction `json:"booked"`
		Pending []Transaction `json:"pending"`
	} `json:"transactions"`
}

type Transaction struct {
	ID          string    `json:"internalTransactionId"`
	BookingDate time.Time `json:"-"`
	ValueDate   time.Time `json:"-"`
	Amount      float64   `json:"-"`
	Currency    string    `json:"-"`
	Memo        string    `json:"remittanceInformationUnstructured"`
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Clone Transaction

	tmp := struct {
		BookingDate       string `json:"bookingDate"`
		ValueDate         string `json:"valueDate"`
		TransactionAmount struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"transactionAmount"`
		*Clone
	}{
		Clone: (*Clone)(t),
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	var err error

	t.BookingDate, err = time.Parse("2006-01-02", tmp.BookingDate)
	if err != nil {
		return err
	}

	t.ValueDate, err = time.Parse("2006-01-02", tmp.ValueDate)
	if err != nil {
		return err
	}

	t.Amount, err = strconv.ParseFloat(tmp.TransactionAmount.Amount, 64)
	t.Currency = tmp.TransactionAmount.Currency

	return nil
}

func (r *Requisition) UnmarshalJSON(data []byte) error {
	type Clone Requisition

	tmp := struct {
		CreatedAt string `json:"created"`
		Link      string `json:"link"`
		*Clone
	}{
		Clone: (*Clone)(r),
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	var err error
	r.CreatedAt, err = time.Parse(time.RFC3339Nano, tmp.CreatedAt)
	if err != nil {
		return err
	}
	r.Link, err = url.Parse(tmp.Link)
	if err != nil {
		return err
	}

	return nil
}

func unwrapError(err error) error {
	if err != nil {
		var e *Error
		if errors.As(err, &e) {
			err = e.Unwrap()
		}
	}
	return err
}
