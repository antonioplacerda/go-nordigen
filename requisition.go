package nordigen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

func (c *Client) CreateRequisition(ctx context.Context, token *Token, redirectURL *url.URL, institutionID string, optional *CreateRequisitionOptional) (*Requisition, error) {
	var res *Requisition
	_, err := c.request(ctx, token, "Create.Requisition").
		WithJSONBody(CreateRequisitionRequest{
			RedirectURL:               redirectURL.String(),
			InstitutionID:             institutionID,
			CreateRequisitionOptional: optional,
		}).
		WithResult(&res).
		Post(ctx, "/api/v2/requisitions/")

	return res, getError(err)
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

func (c *Client) GetRequisition(ctx context.Context, token *Token, requisitionID string) (*Requisition, error) {
	var res *Requisition
	_, err := c.request(ctx, token, "Read.Requisition").
		WithResult(&res).
		Get(ctx, fmt.Sprintf("/api/v2/requisitions/%s", requisitionID))

	return res, getError(err)
}
