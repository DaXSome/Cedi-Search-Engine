package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
)

type LogType = string

const (
	Database LogType = "database"
	Crawler  LogType = "crawler"
	Indexer  LogType = "indexer"
	Sniffer  LogType = "sniffer"
	Error    LogType = "error"
	Utils    LogType = "utils"
	Default  LogType = "default"
)

func Logger(logType LogType, scope string, stmts ...interface{}) {
	defer slog.Close()

	today := time.Now().Format("2006-01-02")

	fmtLogDir := func(scope string) string {
		logDir := filepath.Join("logs", today, fmt.Sprintf("%s.log", scope))

		return logDir
	}

	l := slog.New()

	formatter := slog.NewTextFormatter()
	formatter.SetTemplate("[{{datetime}}] [{{level}}] [{{scope}}] {{message}} {{data}} {{extra}}\n")
	formatter.EnableColor = true

	consoleHandler := handler.NewConsoleHandler(slog.AllLevels)
	consoleHandler.SetFormatter(formatter)

	l.AddHandlers(consoleHandler)

	crawlerFileHandler, err := handler.NewFileHandler(fmtLogDir("crawler"), handler.WithLogfile("crawler.log"))
	crawlerFileHandler.SetFormatter(formatter)
	if err != nil {
		log.Fatalf("Failed to create crawler log file: %v", err)
	}

	indexerFileHandler, err := handler.NewFileHandler(fmtLogDir("indexer"), handler.WithLogfile("indexer.log"))
	indexerFileHandler.SetFormatter(formatter)
	if err != nil {
		log.Fatalf("Failed to create indexer log file: %v", err)
	}

	snifferFileHandler, err := handler.NewFileHandler(fmtLogDir("sniffer"), handler.WithLogfile("sniffer.log"))
	snifferFileHandler.SetFormatter(formatter)
	if err != nil {
		log.Fatalf("Failed to create sniffer log file: %v", err)
	}

	errorFileHandler, err := handler.NewFileHandler(fmtLogDir("error"), handler.WithLogfile("error.log"))
	errorFileHandler.SetFormatter(formatter)
	if err != nil {
		log.Fatalf("Failed to create error log file: %v", err)
	}

	databaseFileHandler, err := handler.NewFileHandler(fmtLogDir("database"), handler.WithLogfile("database.log"))
	databaseFileHandler.SetFormatter(formatter)
	if err != nil {
		log.Fatalf("Failed to create database log file: %v", err)
	}

	utilsFileHandler, err := handler.NewFileHandler(fmtLogDir("utils"), handler.WithLogfile("database.log"))
	utilsFileHandler.SetFormatter(formatter)
	if err != nil {
		log.Fatalf("Failed to create database log file: %v", err)
	}

	switch logType {
	case Crawler:
		l.AddHandlers(crawlerFileHandler)
	case Indexer:
		l.AddHandlers(indexerFileHandler)
	case Sniffer:
		l.AddHandlers(snifferFileHandler)
	case Error:
		l.AddHandlers(errorFileHandler)
	case Database:
		l.AddHandlers(databaseFileHandler)
	case Utils:
		l.AddHandlers(utilsFileHandler)
	}

	if logType == Error {
		l.WithFields(slog.M{
			"scope": scope,
		}).Error(stmts...)
	} else if logType == Default {
		l.WithFields(slog.M{
			"scope": scope,
		}).Debug(stmts...)
	} else {
		l.WithFields(slog.M{
			"scope": scope,
		}).Info(stmts...)
	}
}
