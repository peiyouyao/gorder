package logging

import (
	"context"
	"encoding/json"
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

type ArgFormatter interface {
	FormatArg() (string, error)
}

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
	var (
		str string
		err error
	)
	defer func() {
		if err != nil {
			str = "unsupported type in formatMySQLArg||err=" + err.Error()
		}
	}()
	switch v := arg.(type) {
	default:
		bs, e := json.Marshal(v)
		str, err = string(bs), e
	case ArgFormatter:
		str, err = v.FormatArg()
	}
	return str
}
