package bdc

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// writeToArchive writes the outcome of a function call to the archive file
// By default: archive.txt
func writeToArchive(msg string) error {
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		ioutil.WriteFile(archiveDir, nil, 0666)
	}
	f, err := os.OpenFile(archiveDir, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("Unable to write to archive: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"), msg))
	if err != nil {
		return fmt.Errorf("Unable to write message to archive.\nMessage: %v\nError: %v", msg, err)
	}
	return nil
}

var archiveDir = "./archive.txt"
