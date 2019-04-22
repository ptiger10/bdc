package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type classResponse struct {
	Data []Class `json:"response_data"`
}

// Class is an accounting class in Bill.com (matches QBO class)
type Class struct {
	Entity      string `json:"entity"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
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

// Since returns all Classes updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r classResource) Since(t time.Time, p *Parameters) ([]Class, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(TimeFormat))
	classes, err := r.client.Class.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all Classs updated since %s: %v", t, err)
	}
	return classes, nil
}
