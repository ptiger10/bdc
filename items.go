package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type itemResponse struct {
	Data []Item `json:"response_data"`
}

// Item in Bill.com
type Item struct {
	Entity      string `json:"entity"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
	IsActive    string `json:"isActive"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description"`
}

type itemResource struct {
	resourceFields
}

// All locations
func (r itemResource) All(parameters ...*Parameters) ([]Item, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []Item
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp itemResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}

// Since returns all items updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r itemResource) Since(t time.Time, p *Parameters) ([]Item, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(timeFormat))
	items, err := r.client.Item.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all Items updated since %s: %v", t, err)
	}
	return items, nil
}
