package bdc

import (
	"fmt"
	"time"
)

type customerIdentifier int

// Idenfier options
const (
	ID customerIdentifier = iota
	Name
	AccountNumber
)

// ModifyAllInvoiceDates amends all of a customer's active invoices with invoice dates after today's date
// and non-zero balance by the number of days specified.
// If days is negative, invoice schedule will be moved forward by that many days
// Customer is identified the customer string, which may be in one of three forms:
// bdc.ID: guaranteed to be unique; immutable.
// bdc.Name: guaranteed to unique; mutable.
// bdc.AccountNumber: not guaranteed to be unique; mutable.
// ID is safest but least convenient.
// AccountNumber can be convenient, but multiple people can have same account number.
// Use the form that suits your workflow
func (c *Client) ModifyAllInvoiceDates(identifier string, inputType customerIdentifier, days int) error {
	invoices, err := c.getInvoicesByCustomer(identifier, inputType)
	now := time.Now()
	if err != nil {
		return fmt.Errorf("Unable to modify invoice dates: bad invoice fetch: %v", err)
	}
	for _, invoice := range invoices {
		invoiceDate, err := time.Parse(dateFormat, invoice.InvoiceDate)
		if err != nil {
			return fmt.Errorf("Unable to modify at least one invoice date: bad time parse of InvoiceDate: %v", err)
		}
		dueDate, err := time.Parse(dateFormat, invoice.DueDate)
		if err != nil {
			return fmt.Errorf("Unable to modify at least one invoice date: bad time parse of DueDate: %v", err)
		}
		if (invoice.IsActive == "1") &&
			(invoiceDate.AddDate(0, 0, 1).After(now)) &&
			(invoice.AmountDue > 0) {
			update := Invoice{
				ID:          invoice.ID,
				DueDate:     dueDate.AddDate(0, 0, days).Format(dateFormat),
				InvoiceDate: invoiceDate.AddDate(0, 0, days).Format(dateFormat),
			}
			err := c.Invoice.Update(update)
			if err != nil {
				return fmt.Errorf("Unable to modify at least one invoice date: bad Invoice.Update(): %v", err)
			}
		}

	}
	return nil
}

// DelayInvoicesOneMonth is a convenience method for pushing back all
// active invoices by 30 days
func (c *Client) DelayInvoicesOneMonth(identifier string, inputType customerIdentifier) error {
	err := c.ModifyAllInvoiceDates(identifier, inputType, 30)
	if err != nil {
		return fmt.Errorf("Unable to delay invoices one month: %v", err)
	}
	return err
}

func (c *Client) getInvoicesByCustomer(identifier string, inputType customerIdentifier) ([]Invoice, error) {
	custID, err := c.identifyCustomer(identifier, inputType)
	if err != nil {
		return nil, fmt.Errorf("Unable to get invoices for customer %v: bad identification: %v", identifier, err)
	}
	p := NewParameters()
	p.AddFilter("customerId", "=", custID)
	invoices, err := c.Invoice.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get invoices for customer %v (ID: %v): %v", identifier, custID, err)
	}
	return invoices, nil
}

// convert user-supplied identifier to bill.com ID using different strategies
// if user has supplied a custom inputType (Name, AccountNumber), update that mapping first
// to reduce likelihood of error
func (c *Client) identifyCustomer(identifier string, inputType customerIdentifier) (string, error) {
	var cID string
	var err error
	var ok bool
	var m mapping
	switch inputType {
	case ID:
		cID = identifier
	case Name:
		c.UpdateMappingFile(Customers)
		m, err = getMapping(Customers)
		if err != nil {
			return "", fmt.Errorf("Unable to identify customer: bad name mapping: %v", err)
		}
		cID, ok = m[identifier]
		if !ok {
			return "", fmt.Errorf("Unable to identify customer: name not in map: %v", err)
		}
	case AccountNumber:
		c.UpdateMappingFile(CustomerAccountsID)
		m, err = getMapping(CustomerAccountsID)
		if err != nil {
			return "", fmt.Errorf("Unable to identify customer: bad account number mapping: %v", err)
		}
		cID, ok = m[identifier]
		if !ok {
			return "", fmt.Errorf("Unable to identify customer: account number not in map: %v", err)
		}
	}
	return cID, nil
}
