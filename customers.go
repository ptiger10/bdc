package bdc

import (
	"encoding/json"
	"fmt"
)

type customerResponse struct {
	Data []Customer `json:"response_data"`
}

// Customer in Bill.com
type Customer struct {
	ID           string `json:"id"`
	Entity       string `json:"entity"`
	IsActive     string `json:"isActive"`
	Name         string `json:"name"`
	AccoutNumber string `json:"accNumber"`
	Email        string `json:"email"`
}

type customerResource struct {
	resourceFields
}

// All customers
func (r customerResource) All(parameters ...*Parameters) ([]Customer, error) {
	results := r.client.getAll(r.endpoint, parameters)

	var retList []Customer
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp customerResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}
