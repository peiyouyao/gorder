package persistent

import (
	"context"
	"time"

	"github.com/peiyouyao/gorder/common/util"
	"github.com/sirupsen/logrus"
)

func LogMySQL(ctx context.Context, cmd string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		"mysql_cmd":  cmd,
		"mysql_args": util.FormatArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "mysql_success"
		fields["mysql_cost"] = time.Since(start).Milliseconds()
		fields["mysql_resp"] = resp

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "mysql_error"
			fields["mysql_err"] = (*err).Error()
		}

		logrus.WithContext(ctx).WithFields(fields).Log(level, msg)
	}
}
