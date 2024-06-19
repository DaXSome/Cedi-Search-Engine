package utils

import (
	"log"

	"github.com/fatih/color"
)

func Logger(logType string, stmts ...interface{}) {
	var logger *color.Color

	switch logType {
	case "crawler":
		logger = color.New(color.FgHiCyan)
	case "indexer":
		logger = color.New(color.FgHiGreen)
	case "sniffer":
		logger = color.New(color.FgHiYellow)
	case "error":
		logger = color.New(color.FgHiRed)
	case "database":
		logger = color.New(color.FgHiMagenta)

	default:
		logger = color.New(color.FgHiBlue)
	}

	l := logger.SprintFunc()

	log.Println(l(stmts...))
}
