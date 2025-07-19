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
	setFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel)
	setOutput(logrus.StandardLogger()) // Set output to file and rotate logs
	logrus.AddHook(&traceHook{})
}

// 配置日志格式（本地开发环境支持彩色日志）
func setFormatter(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339, // 标准时间格式
		FieldMap: logrus.FieldMap{ // 重命名字段
			logrus.FieldKeyLevel: "severity", // level -> severity
			logrus.FieldKeyTime:  "time",     // time -> time
			logrus.FieldKeyMsg:   "message",  // msg -> message
		},
	})

	// 如果环境变量 LOCAL_ENV=true，则使用彩色日志格式
	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
		logger.SetFormatter(&prefixed.TextFormatter{
			ForceColors:     true, // 强制使用颜色
			ForceFormatting: true,
			TimestampFormat: time.RFC3339,
		})
	}
}

// 配置日志输出到文件，并分割日志文件
func setOutput(logger *logrus.Logger) {
	var (
		folder        = "./log/"     // 日志文件目录
		logFilepath   = "logs.log"   // 普通日志文件名
		errorFilepath = "errors.log" // 错误日志文件名
	)

	// 确保日志目录存在
	if err := os.MkdirAll(folder, 0750); err != nil && !os.IsExist(err) {
		panic(err)
	}

	// 打开普通日志文件 logs.log（主输出）
	file, err := os.OpenFile(folder+logFilepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	// 打开错误日志文件 errors.log
	_, err = os.OpenFile(folder+errorFilepath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}

	// 设置 logger 的默认输出到 logs.log
	logger.SetOutput(file)

	// 创建普通日志轮转器
	rotateLog, err := rotatelogs.New(
		folder+"%Y%m%d."+logFilepath,             // ➡ 生成 20250719.logs.log 格式
		rotatelogs.WithLinkName(logFilepath),     // ➡ 保留软链接 logs.log 指向当前日志
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 保留最近 7 天日志
		rotatelogs.WithRotationTime(1*time.Hour), // 每小时分割一次日志
	)
	if err != nil {
		panic(err)
	}

	// 创建错误日志轮转器
	rotateError, err := rotatelogs.New(
		folder+"%Y%m%d."+logFilepath,             // ➡ 生成 20250719.errors.log 格式
		rotatelogs.WithLinkName(errorFilepath),   // ➡ 保留软链接 errors.log 指向当前错误日志
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 保留最近 7 天日志
		rotatelogs.WithRotationTime(1*time.Hour), // 每小时分割一次错误日志
	)
	if err != nil {
		panic(err)
	}

	// 将不同日志级别分别写入 rotateLog 和 rotateError
	rotateMap := lfshook.WriterMap{
		logrus.DebugLevel: rotateLog, // Debug 日志写入 20250719.logs.log
		logrus.InfoLevel:  rotateLog,
		logrus.WarnLevel:  rotateLog,
		logrus.ErrorLevel: rotateError, // Error/Fatal/Panic 日志写入 20250719.errors.log
		logrus.FatalLevel: rotateError,
		logrus.PanicLevel: rotateError,
	}

	// 添加 Hook，使用 JSON 格式输出日志（包含时间戳）
	logrus.AddHook(lfshook.NewHook(
		rotateMap,
		&logrus.JSONFormatter{TimestampFormat: time.DateTime},
	))
}

// 带上下文的日志输出函数（支持字段和格式化）
func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Logf(level, format, args...)
}

// 带耗时记录的 Info 日志
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
		entry.Data["trace"] = tracing.TraceID(entry.Context) // 在日志中添加 traceID
		*entry = *entry.WithTime(time.Now())                 // 更新日志时间戳
	}
	return nil
}
