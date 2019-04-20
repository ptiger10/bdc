package bdc

import "fmt"

// reports for common use cases
type reports struct {
	client *Client
}

// OpenInvoices are active and have an amount due greater than 0
func (r reports) OpenInvoices() ([]Invoice, error) {
	p := NewParameters()
	p.AddFilter("isActive", "=", "1")
	p.AddFilter("amountDue", ">", 0)
	p.AddSort("amountDue", 1)
	inv, err := r.client.Invoice.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to complete OpenInvoices report: %v", err)
	}
	return inv, nil
}
