package bdc

import (
	"encoding/json"
	"fmt"
)

type paymentResponse struct {
	Data []PaymentReceived `json:"response_data"`
}

// PaymentReceived in Bill.com, associated with an invoice
type PaymentReceived struct {
	ID            string  `json:"id"`
	CreatedTime   string  `json:"createdTime"`
	UpdatedTime   string  `json:"updatedTime"`
	Amount        float64 `json:"amount"`
	ConvFeeAmount float64 `json:"convFeeAmount"`
	Entity        string  `json:"entity"`
	PaymentDate   string  `json:"paymentDate"`
	IsActive      string  `json:"isActive"`
	CustomerID    string  `json:"customerId"`
	InvoiceNumber string  `json:"invoiceNumber"`
	Description   string  `json:"description"`
	InvoicePays   []struct {
		ID          string  `json:"id"`
		InvoiceID   string  `json:"invoiceId"`
		Amount      float64 `json:"amount"`
		Entity      string  `json:"entity"`
		CreatedTime string  `json:"createdTime"`
		UpdatedTime string  `json:"updatedTime"`
		Description string  `json:"description"`
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
