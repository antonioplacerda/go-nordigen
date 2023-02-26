package nordigen_test

import (
	"context"
	"net/http"
	"net/url"
	"os/exec"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/antonioplacerda/go-nordigen"
)

func TestClient_NewToken(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	client, err := nordigen.New(nil)
	c.Assert(err, qt.IsNil)

	t.Run("NOK - Empty secret_id and secret_key", func(t *testing.T) {
		res, err := client.NewToken(ctx, "", "")
		c.Check(res, qt.IsNil)
		c.Check(err, qt.ErrorMatches, `nordigen error: 400, {"secret_id":\["This field may not be blank."\],"secret_key":\["This field may not be blank."\],"status_code":400}`)
	})

	t.Run("NOK - Empty secret_id", func(t *testing.T) {
		res, err := client.NewToken(ctx, "", "secret_key")
		c.Check(res, qt.IsNil)
		c.Check(err, qt.ErrorMatches, `nordigen error: 400, {"secret_id":\["This field may not be blank."\],"status_code":400}`)
	})

	t.Run("NOK - Empty secret_key", func(t *testing.T) {
		res, err := client.NewToken(ctx, "secret_id", "")
		c.Check(res, qt.IsNil)
		c.Check(err, qt.ErrorMatches, `nordigen error: 400, {"secret_key":\["This field may not be blank."\],"status_code":400}`)
	})

	t.Run("NOK - Invalid pair", func(t *testing.T) {
		res, err := client.NewToken(ctx, "secret_id", "secret_key")
		c.Check(res, qt.IsNil)
		c.Check(err, qt.ErrorMatches, `nordigen error: 401: Authentication failed - No active account found with the given credentials`)
	})

	t.Run("OK - Valid pair", func(t *testing.T) {
		res, err := client.NewToken(ctx, "b721f471-a54f-4af0-98d5-9383f042aec4", "97985fab662556deffe83f7f1ca0617984c1c0ef62a4c7cb5fd9be70ad8dedac01eebd96238efa4d2c8693d8fcdef39cb39e853cdbd908eade707511ca4e0260")
		c.Assert(res, qt.IsNotNil)
		c.Check(res.Access != "", qt.IsTrue)
		c.Check(res.Refresh != "", qt.IsTrue)
		c.Check(res.AccessExpires != 0, qt.IsTrue)
		c.Check(res.RefreshExpires != 0, qt.IsTrue)
		c.Log(res)
		c.Check(err, qt.IsNil)
	})
}

var token = &nordigen.Token{}

func TestClient_ListInstitutions(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	client, err := nordigen.New(nil)
	c.Assert(err, qt.IsNil)

	t.Run("OK", func(t *testing.T) {
		res, err := client.ListInstitutions(ctx, token, "PT")
		c.Check(err, qt.IsNil)
		c.Check(len(res) > 0, qt.IsTrue)
	})

	t.Run("NOK - unauthorized", func(t *testing.T) {
		res, err := client.ListInstitutions(ctx, token, "PT")
		c.Check(err, qt.ErrorIs, nordigen.ErrInvalidToken)
		c.Check(res, qt.IsNil)
	})
}

func TestClient_CreateRequisition(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	client, err := nordigen.New(nil)
	c.Assert(err, qt.IsNil)

	t.Run("OK", func(t *testing.T) {
		link, _ := url.Parse("http://localhost:8080")
		res, err := client.CreateRequisition(ctx, token, link, "BANCOACTIVOBANK_ACTVPTPL", nil)

		go func() {
			err := exec.Command("open", res.Link.String()).Start()
			c.Assert(err, qt.IsNil)
		}()
		ch := make(chan bool, 1)
		go func() {
			http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ch <- true
				_, _ = w.Write([]byte("You can close this window now"))
			}))

			err := http.ListenAndServe(":8080", nil)
			c.Assert(err, qt.IsNil)
		}()
		<-ch

		c.Check(err, qt.IsNil)
		c.Check(res, qt.IsNotNil)
		c.Log(res)
	})

	t.Run("NOK - unauthorized", func(t *testing.T) {
		res, err := client.ListInstitutions(ctx, token, "PT")
		c.Check(err, qt.ErrorIs, nordigen.ErrInvalidToken)
		c.Check(res, qt.IsNil)
	})
}

func TestClient_GetRequisition(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	client, err := nordigen.New(nil)
	c.Assert(err, qt.IsNil)

	t.Run("OK", func(t *testing.T) {
		res, err := client.GetRequisition(ctx, token, "")

		c.Check(err, qt.IsNil)
		c.Check(res, qt.IsNotNil)
		c.Log(res)
	})
}

func TestClient_GetTransactions(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	client, err := nordigen.New(nil)
	c.Assert(err, qt.IsNil)

	t.Run("OK", func(t *testing.T) {
		res, err := client.GetBookedTransactions(ctx, token, "7d7321b6-9c89-42f7-bed1-d502693dc0c3")

		c.Check(err, qt.IsNil)
		c.Check(res, qt.IsNotNil)
		c.Log(res)
	})
}
