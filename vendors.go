package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type vendorResponse struct {
	Data []Vendor `json:"response_data"`
}

// Vendor in Bill.com
type Vendor struct {
	ID           string `json:"id"`
	CreatedTime  string `json:"createdTime"`
	UpdatedTime  string `json:"updatedTime"`
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

// Since returns all vendors updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r vendorResource) Since(t time.Time, p *Parameters) ([]Vendor, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(timeFormat))
	vendors, err := r.client.Vendor.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all vendors updated since %s: %v", t, err)
	}
	return vendors, nil
}
