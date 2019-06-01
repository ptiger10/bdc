package bdc

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
)

// CreateInvoicesFromCSV converts rows in a CSV into Bill.com invoices
// File must match template in "csv_example.csv"
// Best practice is to run c.UpdateInvoiceMappings() prior so that all lookups succeed
func (c *Client) CreateInvoicesFromCSV(path string) error {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read file at %s: %s", path, err)
	}
	reader := csv.NewReader(bytes.NewReader(data))
	records, error := reader.ReadAll()
	if error != nil {
		return fmt.Errorf("error parsing CSV at %s: %s", path, err)
	}

	// Looks for an empty InvoiceNumber or a non-repeated invoice number to demarcate a new invoice
	var invoiceStartLines []int
	if len(records) == 0 {
		invoiceStartLines = append(invoiceStartLines, 0)
	} else {
		for lineNum := 0; lineNum < len(records); lineNum++ {
			row := records[lineNum]
			var repeated bool
			if lineNum+1 == len(records) {
				priorRow := records[lineNum-1]
				repeated = row[1] == priorRow[1]
			} else {
				nextRow := records[lineNum+1]
				repeated = row[1] == nextRow[1]
			}
			header := row[1] == "InvoiceNumber"
			empty := row[1] == ""
			if !header && !empty && !repeated {
				invoiceStartLines = append(invoiceStartLines, lineNum)
			}
		}
	}

	// If nRows = 5 and invoiceStartLines = [1, 3, 5],
	// then linesInInvoice would = [2, 2, 1]
	linesInInvoice := make([]int, len(invoiceStartLines))
	if len(invoiceStartLines) == 0 {
		return fmt.Errorf("no invoices to write")
	}
	for idx := range invoiceStartLines {
		if idx == len(invoiceStartLines)-1 {
			linesInInvoice[idx] = len(records) - invoiceStartLines[idx]
		} else {
			linesInInvoice[idx] = invoiceStartLines[idx+1] - invoiceStartLines[idx]

		}

	}
	for idx, invoiceStartLine := range invoiceStartLines {
		firstRow := records[invoiceStartLine]
		customer := firstRow[0]
		invoiceNumber := firstRow[1]
		dueDate := firstRow[2]
		class := firstRow[3]
		location := firstRow[4]
		var invoiceLineItems []InvoiceLineItem
		for line := invoiceStartLine; line < invoiceStartLine+linesInInvoice[idx]; line++ {
			row := records[line]
			item := row[5]
			amount, err := strconv.ParseFloat(row[6], 8)

			description := row[7]
			li, err := NewInvoiceLineItem("custom", item, amount, description)
			if err != nil {
				return fmt.Errorf("error creating invoice line item on row %v: %v", line, err)
			}
			invoiceLineItems = append(invoiceLineItems, li)
		}
		invoice, err := NewInvoice("custom", customer, invoiceNumber, dueDate, class, location, invoiceLineItems)
		if err != nil {
			return fmt.Errorf("error creating invoice that starts on line %v: %v", invoiceStartLine, err)
		}
		err = c.Invoice.Create(invoice)
		if err != nil {
			return fmt.Errorf("error sending invoice to Bill.com that starts on line %v: %v", invoiceStartLine, err)
		}

	}

	return nil
}
