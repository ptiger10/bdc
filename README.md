[![Go Report Card](https://goreportcard.com/badge/github.com/ptiger10/bdc)](https://goreportcard.com/report/github.com/ptiger10/bdc) [![GoDoc](https://godoc.org/github.com/ptiger10/bdc?status.svg)](https://godoc.org/github.com/ptiger10/bdc) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


# A concurrent Go client for the Bill.com API

**NB: Still under active development. Use at your own risk**

## Logging in to the Bill.com API v2
To make queries to the Bill.com API you must first have an account (username + password), developer key, and organization ID. 
* To get a developer key, contact Bill.com directly (NB: they charge a fee for this) 
* To find your org ID, log in and go to Settings > Your Company > Profile. Look at the end of the URL:  ".../Organization?id=COPY-THIS-VALUE"

 All queries require a `sessionId` which times out after a certain period of activity.

Include your credentials in a json file in the same format as `credentials_example.json`.
By default, it should be named `bdc_credentials.json` though you can set your desired path in the config file (see below). Now you can create a Client with:

`client, err := bdc.GetClient()`

## Get all records
```
client, err := bdc.GetClient()
if err != nil {
    log.Fatal(err)
}
client.Invoice.All()
```

Available options: Vendor, Customer, Invoice, Bill, Location, Class, PaymentMade, PaymentReceived

## Get one record
```
client.Customer.Get("0cu01AAABCDEFGHabc11")
```

## Create one invoice
An invoice is comprised of two parts: a collection of top-level fields that provide invoice metadata, and a list of invoice line items that specify the products/services provided.
```
lineItems := []InvoiceLineItem{
    NewInvoiceLineItem("custom", "Drywall", "50.00", "Drywall for industrials group"),
}
inv, err := NewInvoice("custom", "John Doe", "20190424_doe", "2019-04-24", "Industrials", "San Francisco", lineItems)
if err != nil {
    log.Fatal(err)
}
err = client.Invoice.Create(inv)
if err != nil {
    log.Fatal(err)
}
```

A convenient way to upload multiple invoices simultaneously is to use the `client.CreateInvoicesFromCSV(path)` method.

## Filtering and sorting records
You may filter and sort records by passing a `Parameters` pointer into `client.{Resource}.All(*p)`
```
p := NewParameters()
p.AddFilter("amountDue", ">", 0)
p.AddSort("invoiceDate", 0)
client.Invoice.All(p)
```

## Mappings
Bill.com stores entity IDs as long random strings. In the Bill.com UI, you  interact with these entities as human-readable strings and customize them.

<i>Example</i> <br>Bill.com customer ID: "0cu01AAABCDEFGHabc11"
<br> Human-readable customer name: "John Doe" 

Many workflows, including bulk invoice uploading, require supplying customer information. To simplify this for the client, the `client.FetchAllMappingFiles()` method downloads a local mapping of custom entity identifiers to Bill.com IDs. Each mapping contains a LastUpdated timestamp, so that these maps may be refreshed more quickly in the future with `client.UpdateAllMappingFiles()`.

## History File
Clients commonly want to retain a record of programmatic writes to Bill.com. As long as the `showHistorySelection` field is set to `true` in the config file, a record of all successful create/update actions will be written to a .txt file. Tha path to this file is also specified in the config.

## Config File
The first time a client runs a command from this library, a local file is created at ./.bdc_config.json that stores several default values:
* credentialsFile (string): path to a .json file storing the client's bdc credentials
* mappingsDirectory (string): path to the directory where bdc mappings files will be saved
* historyFile (string): path to a .txt file storing the client's history of upserts into Bill.com
* lastUpdatedFile (string): path to a .txt file storing the last recorded update, for convenience with client.{Resource}.SinceFromFile(...) 
* showHistorySelection (bool): true/false value that determines whether creating/updating invoices will write a confirmation message to a file  on success or simply log the message