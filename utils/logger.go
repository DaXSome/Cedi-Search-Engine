package utils

import (
	dxlogger "github.com/daxsome/daxsome-commons/pkg/logger"
)

type LogType = string

const (
	Database LogType = "database"
	Crawler  LogType = "crawler"
	Indexer  LogType = "indexer"
	Sniffer  LogType = "sniffer"
	Utils    LogType = "utils"

	Error   LogType = "error"
	Default LogType = "default"
)

var logger = dxlogger.NewLogger()

func Logger(logType LogType, scope string, stmts ...interface{}) {
	defer logger.ResetHandlers()

	logFiles := []string{
		"crawler", "indexer", "sniffer", "database", "utils",
	}

	for _, file := range logFiles {
		logger.CreateFileHandler(file)
	}

	logger.Log(logType, scope, stmts...)
}
