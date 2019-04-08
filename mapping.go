package bdc

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

type mapping map[string]string

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

func populateGeneric() {}

// UpdateMappingFile updates the map of {resourceID: value} stored in the mappings/{resource}.json file
// for all active resource items and creates that file if it doesn't exist.
// Also returns inverted mappings for 2-way lookups.
// Options: Locations, Classes, Customers, Vendors
func (c *Client) UpdateMappingFile(resource string) error {
	cleanedInput := handleMappingInput(resource)
	mapping := make(map[string]string)
	mappingInverted := make(map[string]string)
	p := NewParameters()
	p.AddFilter("isActive", "=", "1")
	switch cleanedInput {
	case Locations:
		resp, err := c.Location.All(p)
		if err != nil {
			return fmt.Errorf("Unable to get locations for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	case Classes:
		resp, err := c.Class.All(p)
		if err != nil {
			return fmt.Errorf("Unable to get classes for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	case Customers:
		resp, err := c.Customer.All()
		if err != nil {
			return fmt.Errorf("Unable to get customers for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	case Vendors:
		resp, err := c.Vendor.All()
		if err != nil {
			return fmt.Errorf("Unable to get vendors for mapping: %v", err)
		}
		for _, item := range resp {
			mapping[item.ID] = item.Name
		}
	}
	for k, v := range mapping {
		mappingInverted[v] = k
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
	jsonBlob, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to marshal json for resourceType %v: %v", cleanedInput, err)
	}
	invertedTag := ""
	if inverted {
		invertedTag = "_inverted"
	}
	filePath := path.Join("bdc_mappings", cleanedInput+invertedTag+".json")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Unable to open file for resourceType %v: %v", cleanedInput, err)
	}
	file.Write(jsonBlob)
	return nil
}
