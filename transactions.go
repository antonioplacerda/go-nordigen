package nordigen

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (c *Client) GetBookedTransactions(ctx context.Context, token *Token, accountID string) ([]Transaction, error) {
	var res *TransactionsResponse
	_, err := c.request(ctx, token, "Read.Transactions").
		WithResult(&res).
		Get(ctx, fmt.Sprintf("/api/v2/accounts/%s/transactions/", accountID))

	return res.Transactions.Booked, getError(err)
}

type TransactionsResponse struct {
	Transactions struct {
		Booked  []Transaction `json:"booked"`
		Pending []Transaction `json:"pending"`
	} `json:"transactions"`
}

type Transaction struct {
	ID                    string    `json:"internalTransactionId"`
	BookingDate           time.Time `json:"-"`
	ValueDate             time.Time `json:"-"`
	Amount                float64   `json:"-"`
	Currency              string    `json:"-"`
	AdditionalInformation string    `json:"additionalInformation"`
	CreditorID            string    `json:"creditorId"`
	CreditorName          string    `json:"creditorName"`
	DebtorID              string    `json:"debtorId"`
	DebtorName            string    `json:"debtorName"`
	RemittanceInformation string    `json:"remittanceInformationUnstructured"`
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
