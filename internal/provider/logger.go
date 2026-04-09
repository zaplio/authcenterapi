package provider

import (
	"authcenterapi/internal/provider/dailylogger"
	"authcenterapi/model/constant"
	"authcenterapi/util"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type LogType int

const (
	AppLog = iota
	MongoLog
	PostgresLog
)

type ILogger interface {
	Infof(logType LogType, format string, args ...interface{})
	Infofctx(logType LogType, ctx context.Context, format string, args ...interface{})
	Errorf(logType LogType, format string, args ...interface{})
	Errorfctx(logType LogType, ctx context.Context, addStackTrace bool, format string, args ...interface{})
	Debugf(logType LogType, format string, args ...interface{})
	Debugfctx(logType LogType, ctx context.Context, format string, args ...interface{})
	WithFields(logType LogType, fields logrus.Fields) *logrus.Entry
}

type logrusLogger struct {
	appLog      *logrus.Logger
	mongoLog    *logrus.Logger
	postgresLog *logrus.Logger
}

type CustomFormatter struct {
	TimestampFormat string
	FieldMap        logrus.FieldMap
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	level := entry.Level.String()
	uniqueID := uuid.New().String()

	if reqID, ok := entry.Data["REQUEST_ID"]; ok {
		entry.Data["x-request-id"] = reqID
		entry.Data["uniqueId"] = reqID
		delete(entry.Data, "REQUEST_ID")
	}

	if uniqueId, ok := entry.Data["uniqueId"]; ok {
		uniqueID = uniqueId.(string)
	} else {
		entry.Data["uniqueId"] = uniqueID
	}

	message := entry.Message
	stacktrace := ""
	if stack, ok := entry.Data["stacktrace"]; ok {
		delete(entry.Data, "stacktrace")
		stacktrace = fmt.Sprintf("\nSTACKTRACE:\n%s", stack)
	}

	// Add other fields
	for key, value := range entry.Data {
		if key != "uniqueId" && key != "x-request-id" {
			message = fmt.Sprintf("%s - %s=%v", message, key, value)
		}
	}

	logLine := fmt.Sprintf("%s - %s - %s - %s%s\n", timestamp, strings.ToUpper(level), uniqueID, message, stacktrace)
	return []byte(logLine), nil
}

func NewLogger() ILogger {
	appInfoLogFile := path.Join(util.Configuration.Logger.Dir, "info", fmt.Sprintf("%s.app.info.log", util.Configuration.Logger.FileName))
	appErrorLogFile := path.Join(util.Configuration.Logger.Dir, "error", fmt.Sprintf("%s.app.error.log", util.Configuration.Logger.FileName))

	appLog := logrus.New()
	if util.Configuration.Logger.Level == "debug" {
		appLog.SetLevel(logrus.DebugLevel)
	}
	mongoLog := logrus.New()
	if util.Configuration.Logger.Level == "debug" {
		mongoLog.SetLevel(logrus.DebugLevel)
	}
	postgresLog := logrus.New()
	if util.Configuration.Logger.Level == "debug" {
		postgresLog.SetLevel(logrus.DebugLevel)
	}

	maxAge := util.Configuration.Logger.MaxAge
	maxBackups := util.Configuration.Logger.MaxBackups
	maxSize := util.Configuration.Logger.MaxSize
	compress := util.Configuration.Logger.Compress
	localTime := util.Configuration.Logger.LocalTime

	formatter := &CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	}

	appLog.SetFormatter(formatter)

	appLog.AddHook(&WriterHook{
		Writer: dailylogger.NewDailyRotateLogger(appInfoLogFile, maxSize, maxBackups, maxAge, localTime, compress),
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})

	// Send logs with level higher than warning to stderr
	appLog.AddHook(&WriterHook{
		Writer: dailylogger.NewDailyRotateLogger(appErrorLogFile, maxSize, maxBackups, maxAge, localTime, compress),
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})

	return &logrusLogger{appLog: appLog, mongoLog: mongoLog, postgresLog: postgresLog}
}

func (l *logrusLogger) Infof(logType LogType, format string, args ...interface{}) {
	logger := l.checkType(logType)
	logger.Infof(format, args...)
}
func (l *logrusLogger) Infofctx(logType LogType, ctx context.Context, format string, args ...interface{}) {
	logger := l.checkType(logType)
	requestID, _ := ctx.Value(constant.CtxReqIDKey).(string)
	logger.WithField("REQUEST_ID", requestID).Infof(format, args...)
}

func (l *logrusLogger) Errorf(logType LogType, format string, args ...interface{}) {
	logger := l.checkType(logType)
	logger.Errorf(format, args...)
}

func (l *logrusLogger) Errorfctx(logType LogType, ctx context.Context, addStackTrace bool, format string, args ...interface{}) {
	logger := l.checkType(logType)
	requestID, _ := ctx.Value(constant.CtxReqIDKey).(string)
	log := logger.WithField("REQUEST_ID", requestID)
	if addStackTrace {
		stacktrace := string(debug.Stack())
		log = log.WithField("stacktrace", stacktrace)
	}
	log.Errorf(format, args...)
}

func (l *logrusLogger) Debugf(logType LogType, format string, args ...interface{}) {
	logger := l.checkType(logType)
	logger.Debugf(format, args...)
}

func (l *logrusLogger) Debugfctx(logType LogType, ctx context.Context, format string, args ...interface{}) {
	logger := l.checkType(logType)
	requestID, _ := ctx.Value(constant.CtxReqIDKey).(string)
	logger.WithField("REQUEST_ID", requestID).Debugf(format, args...)
}

func (l *logrusLogger) WithFields(logType LogType, fields logrus.Fields) *logrus.Entry {
	logger := l.checkType(logType)
	return logger.WithFields(fields)
}

func (l *logrusLogger) checkType(logType LogType) *logrus.Logger {
	var logger *logrus.Logger

	if logType == AppLog {
		logger = l.appLog
	} else if logType == MongoLog {
		logger = l.mongoLog
	} else {
		logger = l.postgresLog
	}

	return logger
}

// WriterHook is a hook that writes logs of specified LogLevels to specified Writer
type WriterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func InitLogDir() {
	// workingDirectory, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }

	workingDirectory := util.Configuration.Logger.Dir

	logDirectory := path.Join(workingDirectory)
	if _, err := os.Stat(logDirectory); os.IsNotExist(err) {
		if err := util.CreateDirectory(logDirectory); err != nil {
			panic(err)
		}
	}

	infoLogDirectory := path.Join(logDirectory, "info")
	if _, err := os.Stat(infoLogDirectory); os.IsNotExist(err) {
		if err := util.CreateDirectory(infoLogDirectory); err != nil {
			panic(err)
		}
	}

	errorLogDirectory := path.Join(logDirectory, "error")
	if _, err := os.Stat(errorLogDirectory); os.IsNotExist(err) {
		if err := util.CreateDirectory(errorLogDirectory); err != nil {
			panic(err)
		}
	}

}
