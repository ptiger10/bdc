package bdc

import (
	"encoding/json"
	"fmt"
)

type classResponse struct {
	Data []Class `json:"response_data"`
}

// Class is an accounting class in Bill.com (matches QBO class)
type Class struct {
	Entity      string `json:"entity"`
	IsActive    string `json:"isActive"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description"`
}

type classResource struct {
	resourceFields
}

// All classes
func (r classResource) All(parameters ...*Parameters) ([]Class, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []Class
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp classResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}
