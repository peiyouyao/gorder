package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Method  = "method"
	Args    = "args"
	Cost    = "cost_ms"
	Respose = "response"
	Error   = "error"
)

func WhenMySQL(ctx context.Context, method string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		Method: method,
		Args:   formatMySQLArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "mysql_success"
		fields[Cost] = time.Since(start).Milliseconds()
		fields[Respose] = resp

		if err != nil && *err != nil {
			level, msg = logrus.ErrorLevel, "mysql_error"
			fields[Error] = (*err).Error()
		}
		logrus.WithContext(ctx).WithFields(fields).Log(level, msg)
	}
}

func formatMySQLArgs(args []any) string {
	var argLst []string
	for _, arg := range args {
		argLst = append(argLst, formatMySQLArg(arg))
	}
	return strings.Join(argLst, ",")
}

func formatMySQLArg(arg any) string {
	switch v := arg.(type) {
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("unknown_type(err:%s)", err.Error())
		}
		return string(bytes)
	}
}
