package bdc

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"
)

// CustomerIdentifier is a way that a client can identify a customer
type CustomerIdentifier string

// Idenfier options
// bdc.ID: guaranteed to be unique; immutable.
// bdc.Name: guaranteed to unique; mutable.
// bdc.AccountNumber: not guaranteed to be unique; mutable
// ID is safest but least convenient.
// AccountNumber can be convenient, but multiple people can have same account number.
// Use the form that suits your workflow
const (
	ID            CustomerIdentifier = "id"
	Name                             = "name"
	AccountNumber                    = "account"
)

// Call this only on bill.com-provided date strings that are definitely dates (eg invoice and due dates)
func mustParseDate(date string) time.Time {
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		log.Fatalf("Unable to parse date %s as date: %v", date, err)
	}
	return t
}

// ModifyAllInvoiceDates amends all of a customer's active invoices with invoice dates no more than 1 day old
// and non-zero balance by the number of days specified.
// If days is negative, invoice schedule will be moved forward by that many days
func (c *Client) ModifyAllInvoiceDates(identifier string, inputType CustomerIdentifier, days int) error {
	editableInvoices, err := c.getEditableInvoicesByCustomer(identifier, inputType)
	if err != nil {
		return fmt.Errorf("Unable to modify any invoice dates: %v", err)
	}

	if len(editableInvoices) == 0 {
		return fmt.Errorf("Unable to delay invoice schedule: no editable invoices exist - maybe they are inactive or all in the past?")
	}

	for idx, invoice := range editableInvoices {
		update := Invoice{
			ID:      invoice.ID,
			DueDate: mustParseDate(invoice.DueDate).AddDate(0, 0, days).Format(DateFormat),
		}
		err := c.Invoice.Update(update)
		if err != nil {
			if idx == 0 { // failed on first invoice
				return fmt.Errorf("Unable to modify any invoice dates: %v", err)
			}
			// failed after at least one update succeeded
			return fmt.Errorf("Unable to modify additional invoices: %v", err)
		}
	}
	err = writeToHistory(fmt.Sprintf("Modified all future invoices for customer %v %s by %d days", inputType, identifier, days))
	if err != nil {
		return err
	}
	return nil
}

// Not more than one day old, active, non-zero invoices by customer
// Sorted by due date (ascending order)
func (c *Client) getEditableInvoicesByCustomer(identifier string, inputType CustomerIdentifier) ([]Invoice, error) {
	var editableInvoices []Invoice
	invoices, err := c.getInvoicesByCustomer(identifier, inputType)
	now := time.Now()
	if err != nil {
		return nil, fmt.Errorf("Unable to get editable invoices: %v", err)
	}
	for _, invoice := range invoices {
		if (invoice.IsActive == "1") &&
			(mustParseDate(invoice.InvoiceDate).AddDate(0, 0, 2).After(now)) && // invoiceDate can be today or yesterday to account for time zone variations
			(invoice.AmountDue > 0) {
			editableInvoices = append(editableInvoices, invoice)
		}
	}
	sort.Slice(editableInvoices, func(i, j int) bool {
		return editableInvoices[i].InvoiceDate < editableInvoices[j].InvoiceDate
	})
	return editableInvoices, nil
}

// StretchInvoiceSchedule streches an invoice schedule for a specific customer over a specific number of newMonths
// Sums the amount due on all current active single-line-item invoices with invoice dates no more than 1 day old
// and extends that balance over the number of newMonths provided
// Assumptions: 1 invoice per month, all invoice line items have same value, class, location, and accounting item,
// the last invoice in the series has the latest date
func (c *Client) StretchInvoiceSchedule(identifier string, inputType CustomerIdentifier, newMonths int) error {
	editableInvoices, err := c.getEditableInvoicesByCustomer(identifier, inputType)
	if err != nil {
		return fmt.Errorf("Unable to stretch invoice schedule: %v", err)
	}
	var singleLineInvoices []Invoice
	for _, invoice := range editableInvoices {
		if len(invoice.LineItems) == 1 {
			singleLineInvoices = append(singleLineInvoices, invoice)
		}
	}
	numInvoices := len(singleLineInvoices)
	if numInvoices >= newMonths {
		return fmt.Errorf("Unable to stretch invoice schedule: the newMonths provided (%d) must be greater than the current number of editable invoices (%d)", newMonths, numInvoices)
	}
	if numInvoices == 0 {
		return fmt.Errorf("Unable to stretch invoice schedule: no editable invoices exist - maybe they have multiple line items or are in the past?")
	}

	var totalDue float64
	for _, invoice := range singleLineInvoices {
		totalDue += invoice.AmountDue
	}
	newDuePerInvoice := totalDue / float64(newMonths)

	anchorInvoice := singleLineInvoices[len(singleLineInvoices)-1] // latest invoice date
	newLineItem := anchorInvoice.LineItems[0]
	newLineItem.Price = newDuePerInvoice

	for idx, invoice := range singleLineInvoices {
		err := c.Invoice.Update(Invoice{
			ID:        invoice.ID,
			LineItems: []InvoiceLineItem{newLineItem},
		})
		if err != nil {
			if idx == 0 {
				return fmt.Errorf("Unable to stretch invoice schedule: unable to update any invoices: %v", err)
			}
			return fmt.Errorf("Unable to stretch invoice schedule: unable to update additional invoices: %v", err)
		}
	}

	additionalInvoices := newMonths - numInvoices

	for i := 0; i < additionalInvoices; i++ {
		newDate := mustParseDate(anchorInvoice.DueDate).AddDate(0, i+1, 0).Format(DateFormat) // add one month
		newInvoice, err := NewInvoice(
			"default",
			anchorInvoice.CustomerID,
			anchorInvoice.InvoiceNumber+"_ext"+strconv.Itoa(i+1),
			newDate,
			anchorInvoice.ClassID,
			anchorInvoice.LocationID,
			[]InvoiceLineItem{newLineItem},
		)
		if err != nil {
			return fmt.Errorf("Unable to stretch invoice schedule: unable to populate new invoice for additional months: %v", err)
		}
		err = c.Invoice.Create(newInvoice)
		if err != nil {
			if i == 0 { // failed on first invoice
				return fmt.Errorf("Unable to stretch invoice schedule: unable to create any new invoices: %v", err)
			}
			// failed after at least one update succeeded
			return fmt.Errorf("Unable to stretch invoice schedule: unable to create additional new invoices: %v", err)
		}
	}
	err = writeToHistory(fmt.Sprintf("Stretched invoice schedule for customer %v %s to %d months", inputType, identifier, newMonths))
	if err != nil {
		return err
	}
	return nil

}

func (c *Client) getInvoicesByCustomer(identifier string, inputType CustomerIdentifier) ([]Invoice, error) {
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

// convert client-supplied identifier to bill.com ID using different strategies
// if client has supplied a custom inputType (Name, AccountNumber), update that mapping first
// to reduce likelihood of error
func (c *Client) identifyCustomer(identifier string, inputType CustomerIdentifier) (string, error) {
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
			return "", fmt.Errorf("Unable to identify customer: name not in Customer map: %v", identifier)
		}
	case AccountNumber:
		c.UpdateMappingFile(CustomerAccountsID)
		m, err = getMapping(CustomerAccountsID)
		if err != nil {
			return "", fmt.Errorf("Unable to identify customer: bad account number mapping: %v", err)
		}
		cID, ok = m[identifier]
		if !ok {
			return "", fmt.Errorf("Unable to identify customer: account number not in CustomerAccountsID map: %v", identifier)
		}
	}
	return cID, nil
}
