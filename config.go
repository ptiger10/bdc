package bdc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type defaultConfig interface{}
type pathKey string

var (
	configDefault      defaultConfig = "./.bdc_config.json"
	credentialsDefault               = "./bdc_credentials.json"
	mappingsDefault                  = "./bdc_mappings"
	historyDefault                   = "./bdc_history.txt"
	showHistoryDefault               = true
)

const (
	credentialsFile      pathKey = "bdc_credentialsFile"
	mappingsDirectory            = "bdc_mappingsDir"
	historyFile                  = "bdc_historyFile"
	showHistorySelection         = "bdc_showHistory"
)

var credentialsPath, mappingsDir, historyPath string
var showHistory bool

func init() {
	configDefault := configDefault.(string)
	if _, err := os.Stat(configDefault); os.IsNotExist(err) {
		defaultMap := map[pathKey]defaultConfig{
			credentialsFile:      credentialsDefault,
			mappingsDirectory:    mappingsDefault,
			historyFile:          historyDefault,
			showHistorySelection: showHistoryDefault,
		}
		b, _ := json.MarshalIndent(defaultMap, "", "    ")
		ioutil.WriteFile(configDefault, b, 0666)
	}

	b, _ := ioutil.ReadFile(configDefault)
	var configVars map[pathKey]interface{}
	err := json.Unmarshal(b, &configVars)
	if err != nil {
		log.Fatalf("%v file must have valid JSON: %v", configDefault, err)
	}
	var ok bool
	credentialsPath, ok = configVars[credentialsFile].(string)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type string", credentialsFile, configDefault)
	}
	mappingsDir, ok = configVars[mappingsDirectory].(string)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type string", mappingsDirectory, configDefault)
	}
	historyPath, ok = configVars[historyFile].(string)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type string", historyFile, configDefault)
	}
	showHistory, ok = configVars[showHistorySelection].(bool)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type bool", showHistorySelection, configDefault)
	}

}
