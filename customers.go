package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type customerResponse struct {
	Data []Customer `json:"response_data"`
}

type customerResp struct {
	Data Customer `json:"response_data"`
}

// Customer in Bill.com
type Customer struct {
	ID            string `json:"id"`
	CreatedTime   string `json:"createdTime"`
	UpdatedTime   string `json:"updatedTime"`
	Entity        string `json:"entity"`
	IsActive      string `json:"isActive"`
	Name          string `json:"name"`
	AccountNumber string `json:"accNumber"`
	Email         string `json:"email"`
}

type customerResource struct {
	resourceFields
}

// Get returns a single Customer object
func (r customerResource) Get(id string) (Customer, error) {
	cust, err := r.client.getOne(r.suffix, id)
	if err != nil {
		return Customer{}, fmt.Errorf("Unable to get customer id %v: %v", id, err)
	}
	var goodResp customerResp
	json.Unmarshal(cust, &goodResp)
	return goodResp.Data, nil
}

// All customers
func (r customerResource) All(parameters ...*Parameters) ([]Customer, error) {
	results := r.client.getAll(r.suffix, parameters)

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

// Since returns all customers updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r customerResource) Since(t time.Time, p *Parameters) ([]Customer, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(timeFormat))
	customers, err := r.client.Customer.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all Customers updated since %s: %v", t, err)
	}
	return customers, nil
}
