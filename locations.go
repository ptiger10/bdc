package bdc

import (
	"encoding/json"
	"fmt"
)

type locationResponse struct {
	Data []Location `json:"response_data"`
}

// Location in Bill.com
type Location struct {
	Entity      string `json:"entity"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
	IsActive    string `json:"isActive"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description"`
}

type locationResource struct {
	resourceFields
}

// All locations
func (r locationResource) All(parameters ...*Parameters) ([]Location, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []Location
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp locationResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}
