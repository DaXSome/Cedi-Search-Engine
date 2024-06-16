package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Print error to screen and write to error log file
func HandleErr(err error, logStmt string) bool {
	if err != nil {
		log.Printf("[!] %v", logStmt)

		logFile, err := os.OpenFile("errors.log", os.O_RDWR, 0666)
		if err != nil {
			log.Fatalf("[!!] Couldn't open log file: %v", err)
		}

		logFile.WriteString(fmt.Sprintf("[!] %v: %v => %v\n", time.Now(), logStmt, err))

	}

	return err != nil
}
