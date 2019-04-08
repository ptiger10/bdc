package bdc

import (
	"encoding/json"
	"fmt"
)

type invoiceResponse struct {
	Data []Invoice `json:"response_data"`
}

// Invoice in Bill.com
type Invoice struct {
	ID            string            `json:"id"`
	AmountDue     float64           `json:"amountDue"`
	Amount        float64           `json:"amount"`
	PaymentStatus string            `json:"paymentStatus"`
	Entity        string            `json:"entity"`
	IsActive      string            `json:"isActive"`
	CustomerID    string            `json:"customerId"`
	InvoiceNumber string            `json:"invoiceNumber"`
	InvoiceDate   string            `json:"invoiceDate"`
	DueDate       string            `json:"dueDate"`
	Description   string            `json:"description"`
	LocationID    string            `json:"locationId"`
	ClassID       string            `json:"actgClassId"`
	LineItems     []InvoiceLineItem `json:"invoiceLineItems"`
	ToEmail       bool              `json:"isToBeEmailed"`
}

// InvoiceLineItem on a Bill.com invoice
type InvoiceLineItem struct {
	Entity      string  `json:"entity"`
	ItemID      string  `json:"itemId"`
	Quantity    int     `json:"quantity"`
	Amount      float64 `json:"amount"`
	Price       float64 `json:"price"`
	ClassID     string  `json:"actgClassId"`
	LocationID  string  `json:"locationId"`
	Description string  `json:"description"`
}

type invoiceResource struct {
	resourceFields
}

// All invoices
func (r invoiceResource) All(parameters ...*Parameters) ([]Invoice, error) {
	results := r.client.getAll(r.endpoint, parameters)

	var retList []Invoice
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp invoiceResponse
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}

// Create invoice
func (r invoiceResource) Create(inv Invoice) error {
	error := r.client.createEntity(r.endpoint, inv)
	return error
}

var invoiceSample = `
{
	"response_status" : 0,
	"response_message" : "Success",
	"response_data" : [ {
	  "entity" : "Invoice",
	  "id" : "00abcdefghij123456789",
	  "isActive" : "1",
	  "createdTime" : "2018-12-31T22:00:00.000+0000",
	  "updatedTime" : "2019-01-02T02:01:41.000+0000",
	  "customerId" : "0cuabcdefgh123456789",
	  "invoiceNumber" : "1",
	  "invoiceDate" : "2019-01-01",
	  "dueDate" : "2019-01-01",
	  "glPostingDate" : null,
	  "amount" : 1000.00,
	  "amountDue" : 0.00,
	  "paymentStatus" : "0",
	  "description" : "Invoice for service",
	  "poNumber" : null,
	  "isToBePrinted" : false,
	  "isToBeEmailed" : false,
	  "lastSentTime" : "2019-01-02T02:01:41.000+0000",
	  "itemSalesTax" : "00000000000000000000",
	  "salesTaxPercentage" : 0,
	  "salesTaxTotal" : 0.00,
	  "terms" : "Due upon receipt",
	  "salesRep" : null,
	  "FOB" : null,
	  "shipDate" : null,
	  "shipMethod" : null,
	  "departmentId" : "00000000000000000000",
	  "locationId" : "00000000000000000000",
	  "actgClassId" : "00000000000000000000",
	  "jobId" : "00000000000000000000",
	  "payToBankAccountId" : "00000000000000000000",
	  "payToChartOfAccountId" : "00000000000000000000",
	  "invoiceTemplateId" : "itm01abcdef123456789",
	  "invoiceLineItems" : [ {
		"entity" : "InvoiceLineItem",
		"id" : "00f01abcdef123456789",
		"createdTime" : "2018-12-31T22:00:00.000+0000",
		"updatedTime" : "2019-01-02T02:01:41.000+0000",
		"invoiceId" : "00e01abcdef123456789",
		"itemId" : "0ii01abcdef123456789",
		"quantity" : 1,
		"amount" : 1000.00,
		"price" : 1000.00,
		"serviceDate" : null,
		"ratePercent" : null,
		"chartOfAccountId" : "00000000000000000000",
		"departmentId" : "00000000000000000000",
		"locationId" : "00000000000000000000",
		"actgClassId" : "00000000000000000000",
		"jobId" : "00000000000000000000",
		"description" : "Services rendered",
		"taxable" : true,
		"taxCode" : "Tax"
	  } ]
	} ]
  }
`
