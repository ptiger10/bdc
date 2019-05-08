package bdc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

var configPath = "./.bdc_config.json"

type defaultConfig interface{}
type pathKey string

var (
	credentialsDefault defaultConfig = "./bdc_credentials.json"
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

// SpecifyConfig sets a custom path to the bdc_config.json file
// and overrides "./.bdc_config.json"
func SpecifyConfig(path string) {
	configPath = path
}

// CreateConfig creates a config file with default config values:
//
// * credentialsFile: contains credentials required to authorize a client
//
// * mappingsDir: directory to store mappings between Bill.com IDs and custom client labels, useful in creating new invoices and decoding return values
//
// * historyFile: contains a record of client actions that wrote values to Bill.com
//
// * showHistory: boolean that determines whether to log activity in the historyFile or not
//
// By default, the config file is created at `./.bdc_config.json` To override this location,
// first call bdc.SpecifyConfig(path).
// To supply credentials directly instead of via credentialsFile, call bdc.Login().
func CreateConfig() {
	defaultMap := map[pathKey]defaultConfig{
		credentialsFile:      credentialsDefault,
		mappingsDirectory:    mappingsDefault,
		historyFile:          historyDefault,
		showHistorySelection: showHistoryDefault,
	}
	b, _ := json.MarshalIndent(defaultMap, "", "    ")
	ioutil.WriteFile(configPath, b, 0666)
}

// Runs whenever a Client is created
func loadConfig() error {
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("cannot read config file at %v: %v", configPath, err)
	}
	var configVars map[pathKey]interface{}
	err = json.Unmarshal(b, &configVars)
	if err != nil {
		return fmt.Errorf("%v file must have valid JSON: %v", configPath, err)
	}
	var ok bool
	configDir := filepath.Dir(configPath)
	credentialsPath, ok = configVars[credentialsFile].(string)
	if !ok {
		return fmt.Errorf("value for %q in config file (%q) must be type string", credentialsFile, configPath)
	}
	credentialsPath = filepath.Join(configDir, credentialsPath)

	mappingsDir, ok = configVars[mappingsDirectory].(string)
	if !ok {
		return fmt.Errorf("value for %q in config file (%q) must be type string", mappingsDirectory, configPath)
	}
	mappingsDir = filepath.Join(configDir, mappingsDir)

	historyPath, ok = configVars[historyFile].(string)
	if !ok {
		return fmt.Errorf("value for %q in config file (%q) must be type string", historyFile, configPath)
	}
	historyPath = filepath.Join(configDir, historyPath)

	showHistory, ok = configVars[showHistorySelection].(bool)
	if !ok {
		return fmt.Errorf("Value for %q in config file (%q) must be type bool", showHistorySelection, configPath)
	}
	return nil
}
