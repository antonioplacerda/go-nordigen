package nordigen

import (
	"context"
)

type Institution struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Bic                  string   `json:"bic"`
	TransactionTotalDays string   `json:"transaction_total_days"`
	Countries            []string `json:"countries"`
	LogoURL              string   `json:"logo"`
}

func (c *Client) ListInstitutions(ctx context.Context, token *Token, country string) ([]Institution, error) {
	var res []Institution
	_, err := c.request(ctx, token, "Read.Institutions").
		WithQParam("country", country).
		WithResult(&res).
		Get(ctx, "/api/v2/institutions/")

	return res, getError(err)
}
