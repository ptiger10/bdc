package bdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// convert JSON values into a URL string
func encodeCreateData(c *Client, entity interface{}) io.Reader {
	values := map[string]interface{}{"obj": entity}

	// encode payload as URL
	data := url.Values{}
	jsonValues, _ := json.Marshal(values)
	data.Set("data", string(jsonValues))
	data.Set("sessionId", c.sessionID)
	data.Set("devKey", c.devKey)

	body := strings.NewReader(data.Encode())
	return body
}

func (c *Client) updateEntity(suffix string, entity interface{}) error {
	endpoint := "Crud/Update/" + suffix
	body := encodeCreateData(c, entity)
	_, err := makeRequest(endpoint, body)
	if err != nil {
		return fmt.Errorf("Unable to update item %v at %v: %v", entity, suffix, err)
	}
	return nil

}

// Create entity in Bill.com
func (c *Client) createEntity(suffix string, entity interface{}) error {
	endpoint := "Crud/Create/" + suffix

	body := encodeCreateData(c, entity)
	_, err := makeRequest(endpoint, body)
	if err != nil {
		return fmt.Errorf("Unable to create entity %v at %v: %v", entity, suffix, err)
	}
	return nil
}
