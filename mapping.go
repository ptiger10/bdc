package bdc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type resourceType string

// Resource type options
// For use in mappings
const (
	Locations            resourceType = "Locations"
	Classes                           = "Classes"
	Customers                         = "Customers" // customer name to bill.com id
	Vendors                           = "Vendors"
	Items                             = "Items"
	CustomerAccountsID                = "CustomerAccountsID"   // account number to bill.com id
	CustomerAccountsName              = "CustomerAccountsName" // customer name to account number
)

type mapping map[string]string

// MappingsDir is the directory containing all mappings
const MappingsDir string = "bdc_mappings"

var availableMappings = []resourceType{Locations, Classes, Customers, Vendors, Items, CustomerAccountsID, CustomerAccountsName}

// FetchAllMappingFiles overwrites the map of {resourceID: value} stored in the bdc_mappings/{resource}.json files
// for all active resource items or creates those files if they don't exist.
// For each resource, creates a regular mapping in the form: map[CustomIdentifier]BillDotComIdentifier
// Inactive resourceIDs are ignored.
// The map enables convenient lookups of customer identifiers.
// Every map includes an entry  "*-LastUpdated" with a timestamp of the last time the file was updated
func (c *Client) FetchAllMappingFiles() error {
	log.Printf("Fetching all mapping files and writing to %s/ folder.\nThis may take several moments...",
		MappingsDir)
	for _, resource := range availableMappings {
		err := c.FetchMappingFile(resource)
		if err != nil {
			return fmt.Errorf("Unable to fetch all mappings due to error with %v: %v", resource, err)
		}
	}
	log.Print("All mapping files fetched successfully")
	return nil
}

// FetchMappingFile updates the map of {resourceID: value} stored in the bdc_mappings/{resource}.json file
// for a single resource and creates the file if it doesn't exist.
// Inverts the result returned by the server into a mapping
// in the form: map[CustomIdentifier]BillDotComIdentifier
// Inactive resourceIDs within bill.com are ignored
// Options: Locations, Classes, Customers, Vendors, Items
func (c *Client) FetchMappingFile(resource resourceType) error {
	now := time.Now().UTC()                                        // timestamp at the start of function execution, so no contemporaneous updates are missed in the future
	beginningOfTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC) // in the bill.com world, at least
	mInverted := make(mapping)
	var err error
	p := NewParameters()
	p.AddFilter("isActive", "=", "1")

	m, err := c.fetchMap(resource, beginningOfTime, p)
	if err != nil {
		return fmt.Errorf("Unable to get mapping: %v", err)
	}

	// invert map to make more readable
	for k, v := range m {
		mInverted[v] = k
	}

	if _, err := os.Stat(MappingsDir); os.IsNotExist(err) {
		os.Mkdir(MappingsDir, os.ModePerm)
	}

	err = createOrReplaceMappingFile(mInverted, resource, now)
	if err != nil {
		return fmt.Errorf("Unable to write mapping file for input %v: %v", resource, err)
	}

	return nil
}

// UpdateAllMappingFiles calls UpdateMappingFile() for all available resource types
func (c *Client) UpdateAllMappingFiles() error {
	for _, resource := range availableMappings {
		err := c.UpdateMappingFile(resource)
		if err != nil {
			return fmt.Errorf("Unable to update all mappings due to error with %v: %v", resource, err)
		}
	}
	return nil
}

// UpdateMappingFile only adds or replaces items in the mapping file since it was last updated
// To run this function, you must first have a valid mapping file.
// Create mapping files with c.FetchAllMappingFiles()
func (c *Client) UpdateMappingFile(resource resourceType) error {
	mInverted := make(mapping)
	now := time.Now().UTC() // timestamp at the start of function execution, so no contemporaneous updates are missed in the future
	lastUpdated, err := readLastUpdatedTime(resource)
	if err != nil {
		return fmt.Errorf("Unable to read last updated time for %v: %v", resource, err)
	}
	p := NewParameters()
	p.AddFilter("isActive", "=", "1")
	m, err := c.fetchMap(resource, lastUpdated, p)
	if err != nil {
		return fmt.Errorf("Unable to get mapping: %v", err)
	}

	// invert map to make more readable
	for k, v := range m {
		mInverted[v] = k
	}

	err = updateMappingFile(mInverted, resource, now)
	if err != nil {
		return fmt.Errorf("Unable to update mapping file: %v", err)
	}
	return nil
}

// UpdateInvoiceMappings updates the mapping files in bdc_mappings/
// that assist in creating new invoices and invoice line items
func (c *Client) UpdateInvoiceMappings() error {
	resources := []resourceType{Locations, Classes, Customers, Items}
	for _, resource := range resources {
		err := c.UpdateMappingFile(resource)
		if err != nil {
			return fmt.Errorf("Unable to update all mappings - stopped at %v: %v", resource, err)
		}
	}
	return nil
}

func (c *Client) fetchMap(resource resourceType, t time.Time, p *Parameters) (mapping mapping, err error) {
	switch r := resource; {
	case r == Locations:
		mapping, err = c.locationMap(t, p)
	case r == Classes:
		mapping, err = c.classMap(t, p)
	case r == Customers:
		mapping, err = c.customerMap(t, p)
	case r == Vendors:
		mapping, err = c.vendorMap(t, p)
	case r == Items:
		mapping, err = c.itemMap(t, p)
	case r == CustomerAccountsID:
		mapping, err = c.customerAccountIDMap(t, p)
	case r == CustomerAccountsName:
		mapping, err = c.customerAccountNameMap(t, p)
	default:
		return nil, fmt.Errorf("Unable to find client resource for type %v", resource)
	}
	return
}

// getMapping reads from a file and returns a map for a specified resource
// in form: map[CustomIdentifier]BillDotComIdentifier
func getMapping(resource resourceType) (mapping, error) {
	var m mapping

	fPath := path.Join(MappingsDir, string(resource)+".json")
	b, err := ioutil.ReadFile(fPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %v: %v. Have you run client.FetchAllMappingFiles() yet?", fPath, err)
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, fmt.Errorf("Unable to read JSON at %v: %v", fPath, err)
	}
	return m, nil
}

// returns items mapping for convenience in creating invoice line items
func getItemsMapping() (mapping, error) {
	m, err := getMapping(Items)
	if err != nil {
		return nil, fmt.Errorf("Unable to get convenience mappings for items: %v", err)
	}
	return m, nil
}

// returns a map of resource type names to mappings for customer name, locations, and classes
// for convenience in creating invoices
func getInvoiceCreationMappings() (map[resourceType]mapping, error) {
	var invoiceCreationMappings = []resourceType{Locations, Classes, Customers}
	masterMap := make(map[resourceType]mapping)
	for _, resource := range invoiceCreationMappings {

		m, err := getMapping(resource)
		if err != nil {
			return nil, fmt.Errorf("Unable to get mappings for invoices: %v", err)
		}
		masterMap[resource] = m
	}
	return masterMap, nil
}

// create file if it does not exist or overwrite it completely
func createOrReplaceMappingFile(newMapping mapping, resource resourceType, time time.Time) error {
	newMapping["*-LastUpdated"] = time.Format(TimeFormat)
	jsonBlob, err := json.MarshalIndent(newMapping, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to marshal json for resourceType %v: %v", resource, err)
	}

	filePath := path.Join(MappingsDir, string(resource)+".json")
	err = ioutil.WriteFile(filePath, jsonBlob, 0666)
	if err != nil {
		return fmt.Errorf("Unable to write file for resourceType %v at %v: %v", resource, filePath, err)
	}
	return nil
}

func readLastUpdatedTime(resource resourceType) (time.Time, error) {
	m, err := getMapping(resource)
	if err != nil {
		return time.Time{}, err
	}
	lastUpdated, ok := m["*-LastUpdated"]
	if !ok {
		return time.Time{}, fmt.Errorf("*-LastUpdated tag not in file")
	}
	t, err := time.Parse(TimeFormat, lastUpdated)
	if err != nil {
		return time.Time{}, fmt.Errorf("Last updated time %s not formatted correctly: %v", lastUpdated, err)
	}
	return t, nil
}

func updateMappingFile(updatedMapping mapping, resource resourceType, timestamp time.Time) error {

	// read legacy file
	currentMapping, err := getMapping(resource)
	if err != nil {
		return fmt.Errorf("Unable to update mapping: %v", err)
	}
	for k, v := range updatedMapping {
		currentMapping[k] = v
	}
	// replace legacy file with legacy file plus additions and changes, plus an updated timestamp
	err = createOrReplaceMappingFile(currentMapping, resource, timestamp)
	if err != nil {
		return fmt.Errorf("Unable to update mapping: %v", err)
	}

	return nil
}

func (c *Client) locationMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Location.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get locations for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.ID] = item.Name
	}
	return m, nil
}

func (c *Client) classMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Class.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get classes for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.ID] = item.Name
	}
	return m, nil
}

func (c *Client) customerMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Customer.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get customers for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.ID] = item.Name
	}
	return m, nil
}

func (c *Client) vendorMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Vendor.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get vendors for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.ID] = item.Name
	}
	return m, nil
}

func (c *Client) itemMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Item.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get items for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.ID] = item.Name
	}
	return m, nil
}

func (c *Client) customerAccountIDMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Customer.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get customer accounts for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.ID] = item.AccountNumber
	}
	return m, nil
}

func (c *Client) customerAccountNameMap(t time.Time, p *Parameters) (mapping, error) {
	m := make(mapping)
	resp, err := c.Customer.Since(t, p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get customer accounts for mapping: %v", err)
	}
	for _, item := range resp {
		m[item.AccountNumber] = item.Name
	}
	return m, nil
}
