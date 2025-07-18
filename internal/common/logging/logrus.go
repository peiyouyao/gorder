package logging

import (
	"context"
	"os"
	"strconv"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Using "logging.Infof, Warnf...""
// Or add hook, using "logrus.Infof...""

func Init() {
	SetFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel)
	setOutput(logrus.StandardLogger()) // Set output to file and rotate logs
	logrus.AddHook(&traceHook{})
}

// 配置日志输出到文件并且分割日志
func setOutput(logger *logrus.Logger) {
	var (
		folder        = "./log"
		logFilepath   = "logs.log"
		errorFilepath = "errors.log"
	)

	if err := os.MkdirAll(folder, 0750); err != nil && !os.IsExist(err) {
		panic(err)
	}

	file, err := os.OpenFile(folder+logFilepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	_, err = os.OpenFile(folder+errorFilepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	logger.SetOutput(file)

	rotateLog, err := rotatelogs.New(
		folder+logFilepath+".%Y%m%d",
		rotatelogs.WithLinkName(logFilepath),
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 保留7天的日志
		rotatelogs.WithRotationTime(1*time.Hour), // 每小时分割一次日志
	)
	if err != nil {
		panic(err)
	}

	rotateError, err := rotatelogs.New(
		folder+logFilepath+".%Y%m%d",
		rotatelogs.WithLinkName(errorFilepath),
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 保留7天的日志
		rotatelogs.WithRotationTime(1*time.Hour), // 每小时分割一次日志
	)
	if err != nil {
		panic(err)
	}

	rotateMap := lfshook.WriterMap{
		logrus.DebugLevel: rotateLog,
		logrus.InfoLevel:  rotateLog,
		logrus.WarnLevel:  rotateLog,
		logrus.ErrorLevel: rotateError,
		logrus.FatalLevel: rotateError,
		logrus.PanicLevel: rotateError,
	}

	logrus.AddHook(lfshook.NewHook(rotateMap, &logrus.JSONFormatter{TimestampFormat: time.DateTime}))
}

func SetFormatter(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyMsg:   "message",
		},
	})
	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
		logger.SetFormatter(&prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			TimestampFormat: time.RFC3339,
		})
	}
}

func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Logf(level, format, args...)
}

func InfofWithCost(ctx context.Context, fields logrus.Fields, start time.Time, format string, args ...any) {
	fields[Cost] = time.Since(start).Milliseconds()
	Infof(ctx, fields, format, args...)
}

func Infof(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Infof(format, args...)
}

func Errorf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Errorf(format, args...)
}

func Warnf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Warnf(format, args...)
}

func Panicf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Panicf(format, args...)
}

// traceHook is a logrus hook that adds a trace ID to log entries if the context is not nil.
type traceHook struct{}

func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["trace"] = tracing.TraceID(entry.Context)
		*entry = *entry.WithTime(time.Now()) // OR: entry = entry.WithTime(time.Now())
	}
	return nil
}
