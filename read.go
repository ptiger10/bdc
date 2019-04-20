package bdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

func (c *Client) getOne(suffix string, id string) ([]byte, error) {
	endpoint := "Crud/Read/" + suffix
	body := encodeReadData(c, id)
	resp, err := makeRequest(endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("Unable to get single item %v at %v: %v", id, suffix, err)
	}
	return resp, nil

}

// convert JSON values into a URL string to get one object
func encodeReadData(c *Client, id string) io.Reader {
	values := map[string]interface{}{"id": id}

	// encode payload as URL
	data := url.Values{}
	jsonValues, _ := json.Marshal(values)
	data.Set("data", string(jsonValues))
	data.Set("sessionId", c.sessionID)
	data.Set("devKey", c.devKey)

	body := strings.NewReader(data.Encode())
	return body
}
