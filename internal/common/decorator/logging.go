package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/peiyouyao/gorder/common/logging"
	"github.com/sirupsen/logrus"
)

// impl QueryHandler interface
type queryLoggingDecorator[Q, R any] struct {
	logger  *logrus.Entry
	handler QueryHandler[Q, R]
}

func (d queryLoggingDecorator[Q, R]) Handle(ctx context.Context, query Q) (res R, err error) {
	body, _ := json.Marshal(query)
	fields := logrus.Fields{
		"query":      actionName(query),
		"query_boby": string(body),
	}
	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "query success")
		} else {
			logging.Errorf(ctx, fields, "fail to exec query||err=%v", err)
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
	fields := logrus.Fields{
		"command":      actionName(cmd),
		"command_boby": string(body),
	}
	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "cmd success")
		} else {
			logging.Errorf(ctx, fields, "fail to exec cmd||err=%v", err)
		}
	}()
	res, err = d.handler.Handle(ctx, cmd)
	return
}

func actionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1] // Get the type name without the package
}
