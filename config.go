package bdc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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

// Runs whenever a Client is created
func loadConfig() {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultMap := map[pathKey]defaultConfig{
			credentialsFile:      credentialsDefault,
			mappingsDirectory:    mappingsDefault,
			historyFile:          historyDefault,
			showHistorySelection: showHistoryDefault,
		}
		b, _ := json.MarshalIndent(defaultMap, "", "    ")
		ioutil.WriteFile(configPath, b, 0666)
	}

	b, _ := ioutil.ReadFile(configPath)
	var configVars map[pathKey]interface{}
	err := json.Unmarshal(b, &configVars)
	if err != nil {
		log.Fatalf("%v file must have valid JSON: %v", configPath, err)
	}
	var ok bool
	configDir := filepath.Dir(configPath)
	credentialsPath, ok = configVars[credentialsFile].(string)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type string", credentialsFile, configPath)
	}
	credentialsPath = filepath.Join(configDir, credentialsPath)

	mappingsDir, ok = configVars[mappingsDirectory].(string)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type string", mappingsDirectory, configPath)
	}
	mappingsDir = filepath.Join(configDir, mappingsDir)

	historyPath, ok = configVars[historyFile].(string)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type string", historyFile, configPath)
	}
	historyPath = filepath.Join(configDir, historyPath)

	showHistory, ok = configVars[showHistorySelection].(bool)
	if !ok {
		log.Fatalf("Value for %q in config file (%q) must be type bool", showHistorySelection, configPath)
	}

}
