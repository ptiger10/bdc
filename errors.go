package bdc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type errorResponse struct {
	Status int `json:"response_status"`
	Data   struct {
		Code string `json:"error_code"`
		Msg  string `json:"error_message"`
	} `json:"response_data"`
}

func handleError(r []byte, url string) error {
	var badResp errorResponse
	json.Unmarshal(r, &badResp)
	if badResp.Status == 1 {
		return fmt.Errorf("Unable to perform operation at %s.\tError code: %s\tMessage: %s",
			url, badResp.Data.Code, badResp.Data.Msg)
	}
	return nil
}

func handleErrSlice(errSlice []string) error {
	var err error
	if len(errSlice) > 0 {
		err = errors.New(strings.Join(errSlice, "\n"))
	} else {
		err = nil
	}
	return err
}
