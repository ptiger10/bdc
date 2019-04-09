package bdc

import (
	"fmt"
)

const (
	baseURL    string = "https://api.bill.com/api/v2/"
	loginURL   string = baseURL + "Login.json"
	pageMax    int    = 999
	workersMax int    = 3 // API has a max concurrent thread count of 3
)

const (
	customerSuffix        = "Customer.json"
	vendorSuffix          = "Vendor.json"
	invoiceSuffix         = "Invoice.json"
	billSuffix            = "Bill.json"
	paymentMadeSuffix     = "BillPay.json"
	paymentReceivedSuffix = "ReceivedPay.json"
	locationSuffix        = "Location.json"
	classSuffix           = "ActgClass.json"
	itemSuffix            = "Item.json"
)

type credentials struct {
	UserName string
	Password string
	OrgID    string
	DevKey   string
}

// A Client for making authenticated API calls
type Client struct {
	sessionID       string
	devKey          string
	Customer        customerResource
	Vendor          vendorResource
	Invoice         invoiceResource
	Bill            billResource
	PaymentMade     paymentMadeResource
	PaymentReceived paymentReceivedResource
	Location        locationResource
	Class           classResource
	Item            itemResource
}

type resultError struct {
	result []byte
	page   int
	err    error
}

// Shared fields across all resources
type resourceFields struct {
	suffix string
	client *Client
}

type baseResponse struct {
	Data []map[string]interface{} `json:"response_data"`
}

type resourceType string

// Resource type options
const (
	Locations        resourceType = "Locations"
	Classes                       = "Classes"
	Customers                     = "Customers"
	Vendors                       = "Vendors"
	Invoices                      = "Invoices"
	Bills                         = "Bills"
	BillPayments                  = "BillPayments"
	Payments                      = "Payments"
	Items                         = "Items"
	CustomerAccounts              = "CustomerAccounts"

)

var client = new(Client)

// GetClient returns an authenticated client. Will reuse the existing client if available.
// Must provide the path to a JSON file containing complete Bill.com credentials.
func GetClient(path string) (*Client, error) {
	var creds credentials
	sid, err := login(path, &creds)
	if err != nil {
		return nil, fmt.Errorf("Unable to create new Client: %s", err)
	}
	if client.sessionID == "" {
		client = &Client{sessionID: sid, devKey: creds.DevKey}
	}

	client.Customer = customerResource{resourceFields{suffix: customerSuffix, client: client}}
	client.Vendor = vendorResource{resourceFields{suffix: vendorSuffix, client: client}}
	client.Invoice = invoiceResource{resourceFields{suffix: invoiceSuffix, client: client}}
	client.Bill = billResource{resourceFields{suffix: billSuffix, client: client}}
	client.PaymentMade = paymentMadeResource{resourceFields{suffix: paymentMadeSuffix, client: client}}
	client.PaymentReceived = paymentReceivedResource{resourceFields{suffix: paymentReceivedSuffix, client: client}}
	client.Location = locationResource{resourceFields{suffix: locationSuffix, client: client}}
	client.Class = classResource{resourceFields{suffix: classSuffix, client: client}}
	client.Item = itemResource{resourceFields{suffix: itemSuffix, client: client}}
	return client, nil
}
