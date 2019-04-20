package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type paymentMadeResponse struct {
	Data []PaymentMade `json:"response_data"`
}

// PaymentMade in Bill.com
type PaymentMade struct {
	Entity        string  `json:"entity"`
	CreatedTime   string  `json:"createdTime"`
	UpdatedTime   string  `json:"updatedTime"`
	ID            string  `json:"id"`
	BillID        string  `json:"billId"`
	Name          string  `json:"name"`
	PaymentStatus string  `json:"paymentStatus"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	ProcessDate   string  `json:"processDate"`
}

type paymentMadeResource struct {
	resourceFields
}

// All bill payments
func (r paymentMadeResource) All(parameters ...*Parameters) ([]PaymentMade, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []PaymentMade
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp paymentMadeResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}

// Since returns all payments made since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r paymentMadeResource) Since(t time.Time, p *Parameters) ([]PaymentMade, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(timeFormat))
	payments, err := r.client.PaymentMade.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all payments made since %s: %v", t, err)
	}
	return payments, nil
}

// SinceFileTime returns all payments made (or updated) since the time stored in a text file, eg last_updated.txt.
// File must store a single value formatted according to bdc.timeFormat string
// ie "2006-01-02T15:04:05.999-0700"
// If no additional params to provide, must pass nil explicitly
func (r paymentMadeResource) SinceFileTime(filePath string, params *Parameters) ([]PaymentMade, error) {
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
