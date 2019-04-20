package bdc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type resourceType string

// Resource type options
// For use in mappings
const (
	Locations        resourceType = "Locations"
	Classes                       = "Classes"
	Customers                     = "Customers"
	Vendors                       = "Vendors"
	Items                         = "Items"
	CustomerAccounts              = "CustomerAccounts"
)

type mapping map[string]string

const mappingsDir string = "bdc_mappings"

// getMapping reads from a file and returns a map for a specified resource
// if inverted is false, map will be of form: map[BillDotComIdentifier]CustomIdentifier
// if inverted is true, map will be of form: map[CustomIdentifier]BillDotComIdentifier
func getMapping(resource resourceType, inverted bool) (mapping, error) {
	var m mapping
	var invertedTag string
	if inverted {
		invertedTag = "_inverted"
	}

	fPath := path.Join(mappingsDir, string(resource)+invertedTag+".json")
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

// returns inverted items mapping for convenience in creating invoice line items
func getItemsMapping() (mapping, error) {
	m, err := getMapping(Items, true)
	if err != nil {
		return nil, fmt.Errorf("Unable to get inverted convenience mappings for items: %v", err)
	}
	return m, nil
}

// returns a map of resource type names to inverted mappings for customer name, locations, and classes
// for convenience in creating invoices
func getInvoiceCreationMappings() (map[resourceType]mapping, error) {
	var invoiceCreationMappings = []resourceType{Locations, Classes, Customers}
	masterMap := make(map[resourceType]mapping)
	for _, resource := range invoiceCreationMappings {

		m, err := getMapping(resource, true)
		if err != nil {
			return nil, fmt.Errorf("Unable to get inverted mappings for invoices: %v", err)
		}
		masterMap[resource] = m
	}
	return masterMap, nil
}

var availableMappings = []resourceType{Locations, Classes, Customers, Vendors, Items, CustomerAccounts}

// FetchAllMappingFiles overwrites the map of {resourceID: value} stored in the bdc_mappings/{resource}.json files
// for all active resource items or creates those files if they don't exist.
// For each resource, creates a regular mapping in the form: map[BillDotComIdentifier]CustomIdentifier
// and an inverted mapping of the form: map[CustomIdentifier]BillDotComIdentifier
// Inactive resourceIDs are ignored.
// The inverted map enables convenient lookups of customer identifiers.
// Every map includes an entry  "*-LastUpdated" with a timestamp of the last time the file was updated
func (c *Client) FetchAllMappingFiles() error {
	for _, resource := range availableMappings {
		err := c.FetchMappingFile(resource)
		if err != nil {
			return fmt.Errorf("Unable to fetch all mappings for %v: %v", resource, err)
		}
	}
	return nil
}

// FetchMappingFile updates the map of {resourceID: value} stored in the bdc_mappings/{resource}.json file
// for a single resource and creates the file if it doesn't exist.
// Creates a regular mapping in the form: map[BillDotComIdentifier]CustomIdentifier
// and an inverted mapping of the form: map[CustomIdentifier]BillDotComIdentifier
// Inactive resourceIDs within bill.com are ignored
// Options: Locations, Classes, Customers, Vendors, Items
func (c *Client) FetchMappingFile(resource resourceType) error {
	now := time.Now().UTC()
	mInverted := make(map[string]string)
	var err error
	p := NewParameters()
	p.AddFilter("isActive", "=", "1")

	m, err := c.fetchMap(resource, p)
	if err != nil {
		return fmt.Errorf("Unable to get mapping: %v", err)
	}
	for k, v := range m {
		mInverted[v] = k
	}

	if _, err := os.Stat(mappingsDir); os.IsNotExist(err) {
		os.Mkdir(mappingsDir, os.ModePerm)
	}

	for _, inverted := range []bool{true, false} {
		var newMapping mapping
		if inverted {
			newMapping = mInverted
		} else {
			newMapping = m
		}
		err = createOrReplaceMappingFile(newMapping, resource, inverted)
		if err != nil {
			return fmt.Errorf("Unable to write mapping file for input %v: %v", resource, err)
		}
		err = updateLastUpdated(resource, inverted, now)
		if err != nil {
			return fmt.Errorf("Unable to update last updated time in file: %v", err)
		}
	}

	return nil
}

// create file if it does not exist or overwrite it completely
func createOrReplaceMappingFile(mapping map[string]string, resource resourceType, inverted bool) error {
	jsonBlob, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to marshal json for resourceType %v: %v", resource, err)
	}
	var invertedTag string
	if inverted {
		invertedTag = "_inverted"
	}
	filePath := path.Join(mappingsDir, string(resource)+invertedTag+".json")
	err = ioutil.WriteFile(filePath, jsonBlob, 0666)
	if err != nil {
		return fmt.Errorf("Unable to write file for resourceType %v at %v: %v", resource, filePath, err)
	}
	return nil
}

func readLastUpdatedTime(resource resourceType, inverted bool) (string, error) {
	m, err := getMapping(resource, inverted)
	if err != nil {
		return "", err
	}
	lastUpdated, ok := m["*-LastUpdated"]
	if !ok {
		return "", fmt.Errorf("*-LastUpdated tag not in file")
	}
	return lastUpdated, nil
}

func (c *Client) fetchMap(resource resourceType, p *Parameters) (mapping mapping, err error) {
	switch r := resource; {
	case r == Locations:
		mapping, err = c.locationMap(p)
	case r == Classes:
		mapping, err = c.classMap(p)
	case r == Customers:
		mapping, err = c.customerMap(p)
	case r == Vendors:
		mapping, err = c.vendorMap(p)
	case r == Items:
		mapping, err = c.itemMap(p)
	case r == CustomerAccounts:
		mapping, err = c.customerAccountMap(p)
	default:
		return nil, fmt.Errorf("Unable to find client resource for type %v", resource)
	}
	return
}

// UpdateMappingFile only adds or replaces items in the mapping file since it was last updated
// To run this function, you must first have a valid mapping file.
// Create mapping files with c.FetchAllMappingFiles()
func (c *Client) UpdateMappingFile(resource resourceType) error {
	for _, inverted := range []bool{true, false} {
		now := time.Now().UTC()
		t, err := readLastUpdatedTime(resource, inverted)
		if err != nil {
			return fmt.Errorf("Unable to read last updated time for %v: %v", resource, err)
		}
		p := NewParameters()
		p.AddFilter("isActive", "=", "1")
		p.AddFilter("updatedTime", ">=", t)
		mapping, err := c.fetchMap(resource, p)
		if err != nil {
			return fmt.Errorf("Unable to get mapping: %v", err)
		}

		err = updateMappingFile(mapping, resource, inverted)
		if err != nil {
			return fmt.Errorf("Unable to update mapping file: %v", err)
		}
		err = updateLastUpdated(resource, inverted, now)
		if err != nil {
			return fmt.Errorf("Unable to update last updated time in file: %v", err)
		}
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

func updateLastUpdated(resource resourceType, inverted bool, time time.Time) error {
	m, err := getMapping(resource, inverted)
	if err != nil {
		return fmt.Errorf("Unable to get last updated time: %v", err)
	}
	m["*-LastUpdated"] = time.Format("2006-01-02T15:04:05.999-0700")
	updateMappingFile(m, resource, inverted)
	return nil
}

func updateMappingFile(updatedMapping map[string]string, resource resourceType, inverted bool) error {

	currentMapping, err := getMapping(resource, inverted)
	if err != nil {
		return fmt.Errorf("Unable to update mapping: %v", err)
	}
	for k, v := range updatedMapping {
		currentMapping[k] = v
	}
	err = createOrReplaceMappingFile(currentMapping, resource, inverted)
	if err != nil {
		return fmt.Errorf("Unable to update mapping: %v", err)
	}

	return nil
}

func (c *Client) locationMap(p *Parameters) (mapping, error) {
	mapping := make(map[string]string)
	resp, err := c.Location.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get locations for mapping: %v", err)
	}
	for _, item := range resp {
		mapping[item.ID] = item.Name
	}
	return mapping, nil
}

func (c *Client) classMap(p *Parameters) (mapping, error) {
	mapping := make(map[string]string)
	resp, err := c.Class.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get classes for mapping: %v", err)
	}
	for _, item := range resp {
		mapping[item.ID] = item.Name
	}
	return mapping, nil
}

func (c *Client) customerMap(p *Parameters) (mapping, error) {
	mapping := make(map[string]string)
	resp, err := c.Customer.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get customers for mapping: %v", err)
	}
	for _, item := range resp {
		mapping[item.ID] = item.Name
	}
	return mapping, nil
}

func (c *Client) vendorMap(p *Parameters) (mapping, error) {
	mapping := make(map[string]string)
	resp, err := c.Vendor.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get vendors for mapping: %v", err)
	}
	for _, item := range resp {
		mapping[item.ID] = item.Name
	}
	return mapping, nil
}

func (c *Client) itemMap(p *Parameters) (mapping, error) {
	mapping := make(map[string]string)
	resp, err := c.Item.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get items for mapping: %v", err)
	}
	for _, item := range resp {
		mapping[item.ID] = item.Name
	}
	return mapping, nil
}

func (c *Client) customerAccountMap(p *Parameters) (mapping, error) {
	mapping := make(map[string]string)
	resp, err := c.Customer.All(p)
	if err != nil {
		return nil, fmt.Errorf("Unable to get customer accounts for mapping: %v", err)
	}
	for _, item := range resp {
		mapping[item.ID] = item.AccoutNumber
	}
	return mapping, nil
}
