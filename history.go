package bdc

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// writeToHistory writes the outcome of a function call to the history file
func writeToHistory(msg string) error {
	if !showHistory {
		log.Println(msg)
		return nil
	}
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		ioutil.WriteFile(historyPath, nil, 0666)
	}

	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("Unable to write to history: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"), msg))
	if err != nil {
		return fmt.Errorf("Unable to write message to history.\nMessage: %v\nError: %v", msg, err)
	}
	return nil
}
