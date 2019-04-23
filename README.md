# A concurrent Go client for the Bill.com API

**NB: Still under active development. Use at your own risk**

## Logging in to the Bill.com API v2
To make queries to the Bill.com API you must first have an account (username + password), developer key, and organization ID. 
* To get a developer key, contact Bill.com directly (NB: they charge a fee for this) 
* To find your org ID, log in and go to Settings > Your Company > Profile. Look at the end of the URL:  ".../Organization?id=COPY-THIS-VALUE"

 All queries require a `sessionId` which times out after a certain period of activity.

Include your credentials in a json file in the same format as `credentials_example.json`. Now you can create a Client with:

`c, err := bdc.GetClient()`

## Get all records
```
c, _ := bdc.GetClient()
c.Invoice.All()
```

Available options: Vendor, Customer, Invoice, Bill, Location, Class, PaymentMade, PaymentReceived