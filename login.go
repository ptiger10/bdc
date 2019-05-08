package bdc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type loginResponse struct {
	Data struct {
		SessionID string `json:"sessionId"`
	} `json:"response_data"`
}

// login returns a session id if successful
func login(creds *credentials) (string, error) {
	// Credentials
	f, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return "", fmt.Errorf("Unable to read credentials file (%q) specified in config file (%q): %s", credentialsPath, configPath, err)
	}
	json.Unmarshal(f, creds)
	data := url.Values{}
	data.Set("userName", creds.UserName)
	data.Set("password", creds.Password)
	data.Set("orgId", creds.OrgID)
	data.Set("devKey", creds.DevKey)
	body := strings.NewReader(data.Encode())

	// Request
	resp, err := http.Post(loginURL, "application/x-www-form-urlencoded", body)
	if err != nil {
		return "", fmt.Errorf("Unable to send Post request to %s: %s", loginURL, err)
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Unable to read resp body from %s: %s", loginURL, err)
	}
	// Handling responses
	err = handleError(r, loginURL)
	if err != nil {
		return "", fmt.Errorf("Unable to log in to Bill.com")
	}
	var goodResp loginResponse
	json.Unmarshal(r, &goodResp)

	return goodResp.Data.SessionID, nil
}
