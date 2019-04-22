package bdc

import (
	"encoding/json"
	"fmt"
	"time"
)

type billResponse struct {
	Data []Bill `json:"response_data"`
}

// Bill in Bill.com
type Bill struct {
	Entity        string `json:"entity"`
	CreatedTime   string `json:"createdTime"`
	UpdatedTime   string `json:"updatedTime"`
	IsActive      string `json:"isActive"`
	VendorID      string `json:"vendorId"`
	ID            string `json:"id"`
	InvoiceNumber string `json:"invoiceNumber"`
	InvoiceDate   string `json:"invoiceDate"`
	DueDate       string `json:"dueDate"`
	Description   string `json:"description"`
	LineItems     []struct {
		Entity      string  `json:"entity"`
		Amount      float64 `json:"amount"`
		ItemID      string  `json:"itemId"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"unitPrice"`
		BillID      string  `json:"actgBillId"`
		LocationID  string  `json:"locationId"`
		Description string  `json:"description"`
	} `json:"billLineItems"`
}

type billResource struct {
	resourceFields
}

// All bills
func (r billResource) All(parameters ...*Parameters) ([]Bill, error) {
	results := r.client.getAll(r.suffix, parameters)

	var retList []Bill
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp billResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}

// Since returns all bills updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r billResource) Since(t time.Time, p *Parameters) ([]Bill, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(TimeFormat))
	bills, err := r.client.Bill.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all Bills updated since %s: %v", t, err)
	}
	return bills, nil
}
