package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type paymentResponse struct {
	Data []PaymentReceived `json:"response_data"`
}

// PaymentReceived in Bill.com, associated with an invoice
type PaymentReceived struct {
	Entity      string `json:"entity"`
	ID          string `json:"id"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
	CustomerID  string `json:"customerId"`
	// 0. Paid; 1. Void; 2. Scheduled; 3. Canceled
	Status             string `json:"status"`
	PaymentDate        string `json:"paymentDate"`
	DepositToAccountID string `json:"depositToAccountId"`
	IsOnline           bool   `json:"isOnline"`
	// 0. Cash; 1. Check; 2. CreditCard; 3. ACH; 4. PayPal; 5. Other
	PaymentType   string  `json:"paymentType"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	RefNumber     string  `json:"refNumber"`
	ConvFeeAmount float64 `json:"convFeeAmount"`
	InvoicePays   []struct {
		Entity      string  `json:"entity"`
		ID          string  `json:"id"`
		InvoiceID   string  `json:"invoiceId"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		CreatedTime string  `json:"createdTime"`
		UpdatedTime string  `json:"updatedTime"`
		// 0. Paid; 1. Void; 2. Scheduled; 3. Canceled; 4. Initiated
		Status string `json:"status"`
	} `json:"invoicePays"`
}

type paymentReceivedResource struct {
	resourceFields
}

// All bills
func (r paymentReceivedResource) All(parameters ...*Parameters) ([]PaymentReceived, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []PaymentReceived
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp paymentResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}

// Since returns all payments received since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r paymentReceivedResource) Since(t time.Time, p *Parameters) ([]PaymentReceived, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(TimeFormat))
	payments, err := r.client.PaymentReceived.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all payments received since %s: %v", t, err)
	}
	return payments, nil
}

// SinceFileTime returns all payments updated since the time stored in a text file, eg last_updated.txt.
// File must store a single value formatted according to bdc.TimeFormat string
// ie "2006-01-02T15:04:05.999-0700"
// If no additional params to provide, must pass nil explicitly
func (r paymentReceivedResource) SinceFileTime(filePath string, params *Parameters) ([]PaymentReceived, error) {
	lastUpdated, err := readTimeFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read time from file: %v", err)
	}
	payments, err := r.Since(lastUpdated, params)
	if err != nil {
		return nil, err
	}
	return payments, nil
}
