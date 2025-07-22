package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func Init() {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(true) // 打开 caller 信息
	logger.AddHook(&traceHook{})
	if isLocalEnv() {
		setFormatterLocal(logger)
	} else {
		setFormatterProd(logger)
		setOutput(logger)
	}
}

func setFormatterLocal(logger *logrus.Logger) {
	// 本地开发环境：彩色日志 + caller
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		TimestampFormat:  time.RFC3339,
		FullTimestamp:    true,
		CallerPrettyfier: callerPrettyfier,
	})
}

func setFormatterProd(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:  time.RFC3339,
		CallerPrettyfier: callerPrettyfier,
	})
}

func callerPrettyfier(f *runtime.Frame) (string, string) {
	cwd, _ := os.Getwd()
	relPath := trimCommonPrefix(cwd+"/", f.File)
	return filepath.Base(f.Function), fmt.Sprintf("%s:%d", relPath, f.Line)
}

func trimCommonPrefix(cwd, path string) string {
	i := 0
	for i < len(cwd) && i < len(path) && cwd[i] == path[i] {
		i++
	}
	return path[i:]
}

// 配置日志输出到文件，并分割日志文件
func setOutput(logger *logrus.Logger) {
	var (
		folder        = "./tmp/log/" // 日志文件目录
		logFilepath   = "logs.log"   // 普通日志文件名
		errorFilepath = "errors.log" // 错误日志文件名
	)

	// 确保日志目录存在
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

	// 创建普通日志轮转器
	rotateLog, err := rotatelogs.New(
		filepath.Join(folder, "%Y%m%d."+logFilepath), // 生成 20250721.logs.log 格式
		rotatelogs.WithMaxAge(7*24*time.Hour),        // 保留最近 7 天日志
		rotatelogs.WithRotationTime(1*time.Hour),     // 每小时分割一次日志
	)
	if err != nil {
		panic(err)
	}

	// 创建错误日志轮转器
	rotateError, err := rotatelogs.New(
		filepath.Join(folder, "%Y%m%d."+errorFilepath), // 生成 20250721.errors.log 格式
		rotatelogs.WithMaxAge(7*24*time.Hour),          // 保留最近 7 天错误日志
		rotatelogs.WithRotationTime(1*time.Hour),       // 每小时分割一次错误日志
	)
	if err != nil {
		panic(err)
	}

	// 将不同日志级别分别写入 rotateLog 和 rotateError
	rotateMap := lfshook.WriterMap{
		logrus.DebugLevel: rotateLog, // Debug/Info/Warn 写入 logs.log
		logrus.InfoLevel:  rotateLog,
		logrus.WarnLevel:  rotateLog,
		logrus.ErrorLevel: rotateError, // Error/Fatal/Panic 写入 errors.log
		logrus.FatalLevel: rotateError,
		logrus.PanicLevel: rotateError,
	}

	// 添加 Hook，生产环境用 JSON 格式输出日志
	jsonFormatter := &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
		},
	}

	logrus.AddHook(lfshook.NewHook(rotateMap, jsonFormatter))
}

// traceHook is a logrus hook that adds a trace ID to log entries if the context is not nil.
type traceHook struct{}

func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["trace"] = tracing.TraceID(entry.Context) // 在日志中添加 traceID
		entry.Time = time.Now()                              // 更新日志时间戳
	}
	return nil
}

// 判断是否为本地开发环境
func isLocalEnv() bool {
	isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV"))
	return isLocal
}
