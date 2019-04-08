package bdc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// convert JSON values into a URL string
func encodeCreateData(c *Client, entity interface{}) io.Reader {
	values := map[string]interface{}{"obj": entity}

	// encode payload as URL
	data := url.Values{}
	jsonValues, _ := json.Marshal(values)
	// fmt.Println(string(jsonValues))
	data.Set("data", string(jsonValues))
	data.Set("sessionId", c.sessionID)
	data.Set("devKey", c.devKey)

	body := strings.NewReader(data.Encode())
	return body
}

// Create entity in Bill.com
func (c *Client) createEntity(suffix string, entity interface{}) error {
	endpoint := "Crud/Create/" + suffix

	body := encodeCreateData(c, entity)
	url := baseURL + endpoint
	resp, err := http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return fmt.Errorf("Unable to send Post request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to read resp body from %s: %s", url, err)
	}
	err = handleError(r, url)
	if err != nil {
		return fmt.Errorf("Unable to create invoice: %v", err)
	}
	return nil

}
