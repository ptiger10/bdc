package bdc

import (
	"fmt"
	"log"
	"time"
)

func spin(msg string) {
	log.Print(msg + "\nThis may take several moments...")
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Print("...")
		}
	}()
}
