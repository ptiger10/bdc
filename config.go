package bdc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type defaultPath string
type pathKey string

const (
	configDefault      defaultPath = "./.bdc_config.json"
	credentialsDefault             = "./bdc_credentials.json"
	mappingsDefault                = "./bdc_mappings"
	historyDefault                 = "./bdc_history.txt"
	lastUpdatedDefault             = "./.bdc_last_updated.txt"
)

const (
	credentialsFile      pathKey = "bdc_credentialsFile"
	mappingsDirectory            = "bdc_mappingsDir"
	historyFile                  = "bdc_historyFile"
	lastUpdatedFile              = "bdc_lastUpdatedFile"
	showHistorySelection         = "bdc_showHistory"
)

var credentialsPath, mappingsDir, historyPath, lastUpdatedPath string
var showHistory bool

func init() {
	configDefault := string(configDefault)
	if _, err := os.Stat(configDefault); os.IsNotExist(err) {
		defaultMap := map[string]interface{}{
			string(credentialsFile):      string(credentialsDefault),
			string(mappingsDirectory):    string(mappingsDefault),
			string(historyFile):          string(historyDefault),
			string(lastUpdatedFile):      string(lastUpdatedDefault),
			string(showHistorySelection): true,
		}
		b, _ := json.MarshalIndent(defaultMap, "", "    ")
		ioutil.WriteFile(configDefault, b, 0666)
	}

	b, _ := ioutil.ReadFile(configDefault)
	var configVars map[string]interface{}
	err := json.Unmarshal(b, &configVars)
	if err != nil {
		log.Fatalf("%v file must have valid JSON: %v", configDefault, err)
	}
	credentialsPath = configVars[string(credentialsFile)].(string)
	mappingsDir = configVars[string(mappingsDirectory)].(string)
	historyPath = configVars[string(historyFile)].(string)
	lastUpdatedPath = configVars[string(lastUpdatedFile)].(string)
	showHistory = configVars[string(showHistorySelection)].(bool)

}
