package bdc

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"
)

// Slice of invoices
type invoiceRespList struct {
	Data []Invoice `json:"response_data"`
}

// Single invoice
type invoiceResp struct {
	Data Invoice `json:"response_data"`
}

// Invoice in Bill.com
type Invoice struct {
	ID            string            `json:"id"`
	CreatedTime   string            `json:"createdTime"`
	UpdatedTime   string            `json:"updatedTime"`
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
	results := r.client.getAll(r.suffix, parameters)

	var retList []Invoice
	var errSlice []string
	for _, resp := range results {
		if resp.err != nil {
			errSlice = append(errSlice, fmt.Sprintf("Error on page %v: %v", resp.page, resp.err))
		} else {
			var goodResp invoiceRespList
			json.Unmarshal(resp.result, &goodResp)
			retList = append(retList, goodResp.Data...)
		}
	}
	err := handleErrSlice(errSlice)

	return retList, err
}

// Get returns a single Invoice object
func (r invoiceResource) Get(id string) (Invoice, error) {
	inv, err := r.client.getOne(r.suffix, id)
	if err != nil {
		return Invoice{}, fmt.Errorf("Unable to get invoice id %v: %v", id, err)
	}
	var goodResp invoiceResp
	json.Unmarshal(inv, &goodResp)
	return goodResp.Data, nil
}

// Create invoice
func (r invoiceResource) Create(inv Invoice) error {

	err := r.client.createEntity(r.suffix, inv)
	if err == nil {
		log.Printf("Created Invoice %+v\n", inv)
	}
	return err
}

// Since returns all invoices updated since the time provided.
// If no additional params to provide, must pass nil explicitly
func (r invoiceResource) Since(t time.Time, p *Parameters) ([]Invoice, error) {
	if p == nil {
		p = NewParameters()
	}
	p.AddFilter("updatedTime", ">", t.Format(timeFormat))
	inv, err := r.client.Invoice.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all invoices since %s: %v", t, err)
	}
	return inv, nil
}

// SinceFileTime returns all invoices updated since the time stored in a text file, eg last_updated.txt.
// File must store a single value formatted according to bdc.timeFormat string
// ie "2006-01-02T15:04:05.999-0700"
// If no additional params to provide, must pass nil explicitly
func (r invoiceResource) SinceFileTime(filePath string, params *Parameters) ([]Invoice, error) {
	lastUpdated, err := readTimeFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read time from file: %v", err)
	}
	inv, err := r.Since(lastUpdated, params)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// Update one invoice.
// Supply an Invoice with just the updates you want; all other fields will be preserved.
func (r invoiceResource) Update(updates Invoice) error {
	if updates.ID == "" {
		return fmt.Errorf("Must provide invoice ID to update")
	}
	oldInvoice, err := r.client.Invoice.Get(updates.ID)
	if err != nil {
		return fmt.Errorf("Unable to get invoice %v to run update: %v", updates.ID, err)
	}
	newInvoice := Invoice(oldInvoice)

	valUpdates := reflect.ValueOf(updates)
	for i := 0; i < valUpdates.NumField(); i++ {
		fVal := valUpdates.Field(i)
		fType := fVal.Type()
		fName := valUpdates.Type().Field(i).Name

		var isZero bool
		switch fType.Kind() {
		case reflect.Slice:
			isZero = fVal.Len() == 0
		default:
			isZero = fVal.Interface() == reflect.Zero(fType).Interface()
		}
		if isZero {
			continue
		}

		reflect.ValueOf(&newInvoice).Elem().FieldByName(fName).Set(fVal)
	}
	err = r.client.updateEntity(r.suffix, newInvoice)
	if err == nil {
		log.Printf("Updated Invoice %s with %+v\n", updates.ID, updates)
	}
	return err
}

// NewInvoiceLineItem returns a pointer to a new invoice line item
// Only allows for a quantity of 1 per invoice line item
func NewInvoiceLineItem(itemName string, amount float64, description string) (*InvoiceLineItem, error) {
	maps, err := getItemsMapping()
	if err != nil {
		return nil, fmt.Errorf("Unable to get items mapping: %v", err)
	}
	item, ok := maps[itemName]
	if !ok {
		return nil, fmt.Errorf("Item %v not in mapping. Check file in %v for valid mappings line item", itemName, mappingsDir)
	}
	return &InvoiceLineItem{
		Entity:      "InvoiceLineItem",
		ItemID:      item,
		Amount:      amount,
		Quantity:    1,
		Price:       amount,
		Description: description,
	}, nil
}

// NewInvoice returns a new invoice
// Date must be provided as YYYY-MM-DD
// InvoiceDate and DueDate are set to be equivalent
func NewInvoice(customerName string, invoiceNumber string, dueDate string, className string, locationName string,
	lineItems []*InvoiceLineItem) (Invoice, error) {
	maps, err := getInvoiceCreationMappings()
	if err != nil {
		return Invoice{}, fmt.Errorf("Unable to get convenience mappings to create invoice: %v", err)
	}
	location, ok := maps[Locations][locationName]
	if !ok {
		return Invoice{}, fmt.Errorf("Location %v not in mapping. Check file in %v for valid mappings", locationName, mappingsDir)
	}
	class, ok := maps[Classes][className]
	if !ok {
		return Invoice{}, fmt.Errorf("Class %v not in mapping. Check file in %v for valid mappings", className, mappingsDir)
	}
	customer, ok := maps[Customers][customerName]
	if !ok {
		return Invoice{}, fmt.Errorf("Customer %v not in mapping. Check file in %v for valid mappings", customerName, mappingsDir)
	}

	var amount float64
	var lineItemsCopy []InvoiceLineItem
	for _, lineItem := range lineItems {
		amount += lineItem.Amount
		lineItem.LocationID = location
		lineItem.ClassID = class
		lineItemsCopy = append(lineItemsCopy, *lineItem) // dereference pointers individually
	}

	return Invoice{
		Entity:        "Invoice",
		CustomerID:    customer,
		InvoiceNumber: invoiceNumber,
		InvoiceDate:   dueDate,
		DueDate:       dueDate,
		Amount:        amount,
		AmountDue:     amount, // upon invoice creation, equivalent to amount
		ClassID:       class,
		LocationID:    location,
		ToEmail:       true,

		LineItems: lineItemsCopy,
	}, nil
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
