package logger

import (
	"bytes"
	"electric_payment/config"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type CSVFormatter struct {
	IncludeHeader   bool
	TimestampFormat string
	once            bool
}

func (f *CSVFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	if f.TimestampFormat == "" {
		f.TimestampFormat = "2006-01-02T15:04:05.000Z07:00"
	}

	// Add header row once
	if f.IncludeHeader && !f.once {
		header := []string{"level", "time", "msg", "caller"}
		for k := range entry.Data {
			header = append(header, k)
		}
		b.WriteString(strings.Join(header, ",") + "\n")
		f.once = true
	}

	// Format log fields
	csvFields := []string{
		entry.Level.String(),
		entry.Time.Format(f.TimestampFormat),
		strings.ReplaceAll(entry.Message, ",", ";"), // sanitize commas
	}

	// Add shortened caller (remove current working dir)
	caller := ""
	if entry.HasCaller() {
		wd, err := os.Getwd()
		if err == nil {
			cleanPath := strings.TrimPrefix(entry.Caller.File, wd+"/")
			caller = fmt.Sprintf("%s:%d", cleanPath, entry.Caller.Line)
		} else {
			caller = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		}
	}
	csvFields = append(csvFields, caller)

	// Add extra fields
	for _, v := range entry.Data {
		csvFields = append(csvFields, fmt.Sprint(v))
	}

	b.WriteString(strings.Join(csvFields, ",") + "\n")
	return b.Bytes(), nil
}

// InitLogger initializes the Logrus logger with daily log rotation
func InitLogrus() {
	appLogDir := config.GetConfig().App.LogDir
	if err := os.MkdirAll(appLogDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	logPath := filepath.Join(appLogDir, config.GetConfig().App.SystemLogFilename)

	logLevel := config.GetConfig().App.LogLevel
	switch strings.ToLower(logLevel) {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn", "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.TraceLevel)
	}

	// Set log format from environment (default to JSON)
	logFormat := config.GetConfig().App.LogFormat
	switch strings.ToLower(logFormat) {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&CSVFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000 MST",
		})
	}

	// Setup daily log rotation
	logrus.Info("✍🏻 Log file path: ", logPath)
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   logPath, // Log file name
		MaxSize:    50,      // Max size in MB before rotation
		MaxAge:     30,      // Days to keep old log files
		MaxBackups: 20,      // Maximum number of backup logs
		Compress:   true,    // Compress old log files -> e.g. *.tar.gz
	})
	logrus.SetReportCaller(true)
	logrus.Info("🟢 Logger initialized successfully")
}
