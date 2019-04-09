package bdc

import (
	"encoding/json"
	"fmt"
)

type paymentMadeResponse struct {
	Data []PaymentMade `json:"response_data"`
}

// PaymentMade in Bill.com
type PaymentMade struct {
	Entity        string  `json:"entity"`
	ID            string  `json:"id"`
	BillID        string  `json:"billId"`
	Name          string  `json:"name"`
	PaymentStatus string  `json:"paymentStatus"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	ProcessDate   string  `json:"processDate"`
	CreatedTime   string  `json:"createdTime"`
	UpdatedTime   string  `json:"updatedTime"`
}

type paymentMadeResource struct {
	resourceFields
}

// All bills
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
