package bdc

import (
	"encoding/json"
	"fmt"
	"time"
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

// Since returns all locations updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r locationResource) Since(t time.Time, p *Parameters) ([]Location, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(timeFormat))
	locations, err := r.client.Location.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all Locations updated since %s: %v", t, err)
	}
	return locations, nil
}
