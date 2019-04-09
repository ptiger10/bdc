package bdc

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
)

// CreateInvoiceFromCSV converts rows in a CSV into Bill.com invoices
// File must match template in "csv_example.csv"
func CreateInvoiceFromCSV(path string) error {
	c, err := GetClient("credentials.json")
	if err != nil {
		return fmt.Errorf("Error creating client: %s", err)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Unable to read file at %s: %s", path, err)
	}
	reader := csv.NewReader(bytes.NewReader(data))
	records, error := reader.ReadAll()
	if error != nil {
		return fmt.Errorf("Error parsing CSV at %s: %s", path, err)
	}

	// Looks for an empty InvoiceNumber to demarcate a new invoice
	var invoiceStartLines []int
	for lineNum, row := range records {
		if row[1] != "InvoiceNumber" && row[1] != "" {
			invoiceStartLines = append(invoiceStartLines, lineNum)
		}
	}

	var lineOffsets []int
	if len(invoiceStartLines) == 0 {
		return fmt.Errorf("No invoices to write")
	}
	for idx, val := range invoiceStartLines {
		if idx == len(invoiceStartLines)-1 {
			lineOffsets = append(lineOffsets, len(records)-val)
		} else {
			lineOffsets = append(lineOffsets, invoiceStartLines[idx+1]-val)

		}

	}
	for idx, invoiceStartLine := range invoiceStartLines {
		firstRow := records[invoiceStartLine]
		customer := firstRow[0]
		invoiceNumber := firstRow[1]
		dueDate := firstRow[2]
		class := firstRow[3]
		location := firstRow[4]
		var invoiceLineItems []*InvoiceLineItem
		for offset := invoiceStartLine; offset < invoiceStartLine+lineOffsets[idx]; offset++ {
			row := records[offset]
			item := row[5]
			amount, err := strconv.ParseFloat(row[6], 8)

			description := row[7]
			li, err := NewInvoiceLineItem(item, amount, description)
			if err != nil {
				return fmt.Errorf("Error making invoice line item on row %v: %v", offset, err)
			}
			invoiceLineItems = append(invoiceLineItems, li)
		}
		invoice, err := NewInvoice(customer, invoiceNumber, dueDate, class, location, invoiceLineItems)
		if err != nil {
			return fmt.Errorf("Error populating invoice that starts on line %v: %v", invoiceStartLine, err)
		}
		err = c.Invoice.Create(invoice)
		if err != nil {
			return fmt.Errorf("Error sending invoice to Bill.com that starts on line %v: %v", invoiceStartLine, err)
		}

	}

	return nil
}
