package bdc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type mapping map[string]string

const mappingsDir string = "bdc_mappings"

var availableMappings = []resourceType{Locations, Classes, Customers, Vendors, Items, CustomerAccounts}
var invoiceCreationMappings = []resourceType{Locations, Classes, Customers}

func handleMappingInput(input string) resourceType {
	output := strings.Title(input)
	output = strings.TrimSpace(output)
	if !strings.HasSuffix(output, "s") {
		output += "s"
	}
	if strings.HasSuffix(output, "ss") { // Handle "Class" input
		output += "es"
	}
	return resourceType(output)
}

// getMapping returns a map for a specified resource
func getMapping(resource resourceType, inverted bool) (mapping, error) {
	var m mapping
	var invertedTag string
	if inverted {
		invertedTag = "_inverted"
	}

	fPath := path.Join(mappingsDir, string(resource)+invertedTag+".json")
	b, err := ioutil.ReadFile(fPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %v: %v. Have you run client.UpdateMappingFileAll() yet?", fPath, err)
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, fmt.Errorf("Unable to read JSON at %v: %v", fPath, err)
	}
	return m, nil
}

func getItemsMapping() (mapping, error) {
	m, err := getMapping(Items, true)
	if err != nil {
		return nil, fmt.Errorf("Unable to get inverted convenience mappings for items: %v", err)
	}
	return m, nil
}

// returns inverted mappings for customer name, locations, and classes
func getInvoiceCreationMappings() (map[resourceType]mapping, error) {
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

// UpdateMappingFileAll updates the map of {resourceID: value} stored in the mappings/{resource}.json files
// for all active resource items and creates those files if they don't exist.
// Also returns inverted mappings for 2-way lookups.
func (c *Client) UpdateMappingFileAll() error {
	for _, resource := range availableMappings {
		err := c.updateMappingFile(string(resource))
		if err != nil {
			return fmt.Errorf("Unable to update all mappings due to error with %v: %v", resource, err)
		}
	}
	return nil
}

// updateMappingFile updates the map of {resourceID: value} stored in the mappings/{resource}.json file
// for a single resource and creates the file if it doesn't exist.
// Also returns inverted mappings for 2-way lookups.
// Options: Locations, Classes, Customers, Vendors
func (c *Client) updateMappingFile(resource string) error {
	cleanedInput := handleMappingInput(resource)
	mapping := make(map[string]string)
	mappingInverted := make(map[string]string)
	p := NewParameters()
	p.AddFilter("isActive", "=", "1")
	switch r := cleanedInput; {
	case r == Locations:
		resp, err := c.Location.All(p)
		if err != nil {
			return fmt.Errorf("Unable to get locations for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	case r == Classes:
		resp, err := c.Class.All(p)
		if err != nil {
			return fmt.Errorf("Unable to get classes for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	case r == Customers:
		resp, err := c.Customer.All()
		if err != nil {
			return fmt.Errorf("Unable to get customers for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	case r == Vendors:
		resp, err := c.Vendor.All()
		if err != nil {
			return fmt.Errorf("Unable to get vendors for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}

	case r == Items:

		resp, err := c.Item.All()
		if err != nil {
			return fmt.Errorf("Unable to get items for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}

	case r == CustomerAccounts:
		resp, err := c.Customer.All()
		if err != nil {
			return fmt.Errorf("Unable to get items for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.AccoutNumber
		}

	}
	for k, v := range mapping {
		mappingInverted[v] = k
	}

	if _, err := os.Stat(mappingsDir); os.IsNotExist(err) {
		os.Mkdir(mappingsDir, os.ModePerm)
	}

	err := writeToMappingFile(mapping, string(cleanedInput), false)
	if err != nil {
		return fmt.Errorf("Unable to write to mapping file for input %v: %v", resource, err)
	}
	err = writeToMappingFile(mappingInverted, string(cleanedInput), true)
	if err != nil {
		return fmt.Errorf("Unable to write to mapping file for input %v: %v", resource, err)
	}
	return err
}

func writeToMappingFile(mapping map[string]string, cleanedInput string, inverted bool) error {
	mapping["*-LastUpdated"] = time.Now().UTC().Format("2006-01-02T15:04:05.999-0700")
	jsonBlob, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to marshal json for resourceType %v: %v", cleanedInput, err)
	}
	var invertedTag string
	if inverted {
		invertedTag = "_inverted"
	}
	filePath := path.Join(mappingsDir, cleanedInput+invertedTag+".json")
	err = ioutil.WriteFile(filePath, jsonBlob, 0666)
	if err != nil {
		return fmt.Errorf("Unable to write file for resourceType %v at %v: %v", cleanedInput, filePath, err)
	}
	return nil
}
