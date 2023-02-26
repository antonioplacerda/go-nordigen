package nordigen

import (
	"context"

	"github.com/antonioplacerda/requests"
)

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
	_, err := requests.NewRequest(c.baseURL, c.hc, "application/json").
		WithJSONBody(NewTokenRequest{
			SecretID:  secretID,
			SecretKey: secretKey,
		}).
		WithResult(&res).
		Post(ctx, "/api/v2/token/new/")

	// {"secret_id":["This field may not be blank."],"secret_key":["This field may not be blank."],"status_code":400}
	return res, getError(err)
}
