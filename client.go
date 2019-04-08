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
	customerEndpoint        = "Customer.json"
	vendorEndpoint          = "Vendor.json"
	invoiceEndpoint         = "Invoice.json"
	billEndpoint            = "Bill.json"
	paymentMadeEndpoint     = "BillPay.json"
	paymentReceivedEndpoint = "ReceivedPay.json"
	locationEndpoint        = "Location.json"
	classEndpoint           = "ActgClass.json"
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
}

type resultError struct {
	result []byte
	page   int
	err    error
}

// Shared fields across all resources
type resourceFields struct {
	endpoint string
	client   *Client
}

type baseResponse struct {
	Data []map[string]interface{} `json:"response_data"`
}

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

	client.Customer = customerResource{resourceFields{endpoint: customerEndpoint, client: client}}
	client.Vendor = vendorResource{resourceFields{endpoint: vendorEndpoint, client: client}}
	client.Invoice = invoiceResource{resourceFields{endpoint: invoiceEndpoint, client: client}}
	client.Bill = billResource{resourceFields{endpoint: billEndpoint, client: client}}
	client.PaymentMade = paymentMadeResource{resourceFields{endpoint: paymentMadeEndpoint, client: client}}
	client.PaymentReceived = paymentReceivedResource{resourceFields{endpoint: paymentReceivedEndpoint, client: client}}
	client.Location = locationResource{resourceFields{endpoint: locationEndpoint, client: client}}
	client.Class = classResource{resourceFields{endpoint: classEndpoint, client: client}}
	return client, nil
}
