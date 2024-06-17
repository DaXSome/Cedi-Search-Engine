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
		fullLog := fmt.Sprintf("[!] %v: %v => %v\n", time.Now(), logStmt, err)
		log.Printf(fullLog)

		logFile, err := os.OpenFile("errors.log", os.O_RDWR, 0666)
		if err != nil {
			log.Fatalf("[!!] Couldn't open log file: %v", err)
		}

		logFile.WriteString(fullLog)

	}

	return err != nil
}
