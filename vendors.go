package bdc

import (
	"encoding/json"
	"fmt"
)

type vendorResponse struct {
	Data []Vendor `json:"response_data"`
}

// Vendor in Bill.com
type Vendor struct {
	ID           string `json:"id"`
	Entity       string `json:"entity"`
	Name         string `json:"name"`
	IsActive     string `json:"isActive"`
	AccoutNumber string `json:"accNumber"`
	Email        string `json:"email"`
}

type vendorResource struct {
	resourceFields
}

// All vendors
func (r vendorResource) All(parameters ...*Parameters) ([]Vendor, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []Vendor
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp vendorResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}
