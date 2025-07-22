package logging

import (
	"context"
	"time"

	"github.com/peiyouyao/gorder/common/util"
	"github.com/sirupsen/logrus"
)

func LoggingWithCost(ctx context.Context, method string, fields logrus.Fields) (dlog func(resp any, err error)) {
	start := time.Now()

	// 拷贝字段，避免外部修改污染
	fs := make(logrus.Fields, len(fields))
	for k, v := range fields {
		fs[k] = v
	}

	dlog = func(res any, err error) {
		if rec := recover(); rec != nil {
			fs["panic"] = rec
			fs["time_cost"] = time.Since(start)
			logrus.WithContext(ctx).WithFields(fs).Fatalf("%s_panic", method)
			panic(rec) // 继续抛出 panic
		}

		fs["res"] = res
		fs["time_cost"] = time.Since(start)
		if err == nil {
			logrus.WithContext(ctx).WithFields(fs).Infof("%s_success", method)
		} else {
			fs["err"] = err.Error()
			logrus.WithContext(ctx).WithFields(fs).Errorf("%s_fail", method)
		}
	}
	return
}

func WhenEventPublish(ctx context.Context, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		"args": util.FormatArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "_mq_publish_success"
		fields["publish_time_cost"] = time.Since(start).Milliseconds()
		fields["publish_resp"] = resp

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "_mq_publish_failed"
			fields["publish_error"] = (*err).Error()
		}

		logrus.WithContext(ctx).WithFields(fields).Log(level, msg)
	}
}
