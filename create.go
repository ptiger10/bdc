package bdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// bill.com cannot accept an entity with a CreatedTime or UpdatedTime field
func removeTimestamps(entity interface{}) (interface{}, error) {
	originalJSON, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("Unable to remove timestamps data: unable to marshal original json: %v", err)
	}
	var entityMap map[string]interface{}
	err = json.Unmarshal(originalJSON, &entityMap)
	if err != nil {
		return nil, fmt.Errorf("Unable to remove timestamps: unable to unmarshal original json: %v", err)
	}
	delete(entityMap, "createdTime")
	delete(entityMap, "updatedTime")
	return entityMap, nil
}

// convert JSON values into a URL string
func encodeCreateData(c *Client, entity interface{}) (io.Reader, error) {
	entity, err := removeTimestamps(entity)
	if err != nil {
		return nil, fmt.Errorf("Unable to encode data: %v", err)
	}
	values := map[string]interface{}{"obj": entity}

	// encode payload as URL
	data := url.Values{}
	jsonValues, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("Unable to encode data: unable to marshal cleaned json: %v", err)
	}
	data.Set("data", string(jsonValues))
	data.Set("sessionId", c.sessionID)
	data.Set("devKey", c.devKey)

	body := strings.NewReader(data.Encode())
	return body, nil
}

func (c *Client) updateEntity(suffix string, entity interface{}) (string, error) {
	endpoint := "Crud/Update/" + suffix
	body, err := encodeCreateData(c, entity)
	if err != nil {
		return "", fmt.Errorf("Unable to update entity: %v", err)
	}
	r, err := makeRequest(endpoint, body)
	if err != nil {
		return "", fmt.Errorf("Unable to update item at %v: %v", suffix, err)
	}
	var resp confirmationResponse
	json.Unmarshal(r, &resp)
	confirmation := fmt.Sprintf("%v", resp.Data)
	return confirmation, nil

}

// Create entity in Bill.com
func (c *Client) createEntity(suffix string, entity interface{}) (string, error) {
	endpoint := "Crud/Create/" + suffix

	body, err := encodeCreateData(c, entity)
	if err != nil {
		return "", fmt.Errorf("Unable to update entity: %v", err)
	}
	r, err := makeRequest(endpoint, body)
	if err != nil {
		return "", fmt.Errorf("Unable to create entity at %v: %v", suffix, err)

	}
	var resp confirmationResponse
	json.Unmarshal(r, &resp)
	confirmation := fmt.Sprintf("%v", resp.Data)
	return confirmation, nil
}
