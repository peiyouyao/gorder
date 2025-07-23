package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// impl QueryHandler interface
type queryLoggingDecorator[Q, R any] struct {
	logger  *logrus.Entry
	handler QueryHandler[Q, R]
}

func (d queryLoggingDecorator[Q, R]) Handle(ctx context.Context, query Q) (res R, err error) {
	body, _ := json.Marshal(query)
	fs := logrus.Fields{
		"query":      actionName(query),
		"query_body": string(body),
	}
	defer func() {
		if err == nil {
			logrus.WithContext(ctx).WithFields(fs).Info("Query ok")
		} else {
			fs["query_err"] = err.Error()
			logrus.WithContext(ctx).WithFields(fs).Error("Query fail")
		}
	}()
	res, err = d.handler.Handle(ctx, query)
	return
}

// impl CommandHandler interface
type commandLoggingDecorator[C, R any] struct {
	logger  *logrus.Entry
	handler CommandHandler[C, R]
}

func (d commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (res R, err error) {
	body, _ := json.Marshal(cmd)
	fs := logrus.Fields{
		"command":      actionName(cmd),
		"commond_body": string(body),
	}
	defer func() {
		if err == nil {
			logrus.WithContext(ctx).WithFields(fs).Info("Command ok")
		} else {
			fs["command_err"] = err.Error()
			logrus.WithContext(ctx).WithFields(fs).Error("Command fail")
		}
	}()
	res, err = d.handler.Handle(ctx, cmd)
	return
}

func actionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1] // Get the type name without the package
}
