package utils

import (
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

	LogFormat = "[{{datetime}}] [{{level}}] [{{scope}}] {{message}} {{data}} {{extra}}\n"
)

func Logger(logType LogType, scope string, stmts ...interface{}) {
	defer slog.Close()

	today := time.Now().Format("2006-01-02")
	logDir := filepath.Join("logs", today)

	createFileHandler := func(fileName string) slog.Handler {
		logPath := filepath.Join(logDir, fileName)
		h, err := handler.NewFileHandler(logPath)
		if err != nil {
			log.Fatalf("Failed to create log file %s: %v", logPath, err)
		}

		fileFormatter := slog.NewTextFormatter()
		fileFormatter.SetTemplate(LogFormat)

		h.SetFormatter(fileFormatter)
		return h
	}

	getLogLevel := func() slog.Level {
		switch logType {
		case Error:
			return slog.ErrorLevel

		case Default:
			return slog.DebugLevel

		default:
			return slog.InfoLevel
		}

	}

	consoleFormatter := slog.NewTextFormatter()
	consoleFormatter.SetTemplate(LogFormat)
	consoleFormatter.EnableColor = true

	consoleHandler := handler.NewConsoleHandler(slog.AllLevels)
	consoleHandler.SetFormatter(consoleFormatter)

	l := slog.New()
	l.AddHandlers(consoleHandler)

	handlers := map[LogType]slog.Handler{
		Crawler:  createFileHandler("crawler.log"),
		Indexer:  createFileHandler("indexer.log"),
		Sniffer:  createFileHandler("sniffer.log"),
		Error:    createFileHandler("error.log"),
		Database: createFileHandler("database.log"),
		Utils:    createFileHandler("utils.log"),
	}

	if h, ok := handlers[logType]; ok {
		l.AddHandlers(h)
	}

	l.WithFields(slog.M{
		"scope": scope,
	}).Log(getLogLevel(), stmts...)
}
