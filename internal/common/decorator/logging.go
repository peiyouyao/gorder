package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   QueryHandler[C, R]
}

// impl QueryHandler interface
func (d queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := d.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": string(body),
	})
	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logger.Info("Query executed successfully")
		} else {
			logger.Error("Failed to execute query", err)
		}
	}()
	res, err := d.base.Handle(ctx, cmd)
	return res, err
}

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   CommandHandler[C, R]
}

func (d commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := d.logger.WithFields(logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": string(body),
	})
	logger.Debug("Executing command")
	defer func() {
		if err == nil {
			logger.Info("Command executed successfully")
		} else {
			logger.Error("Failed to execute command", err)
		}
	}()
	res, err := d.base.Handle(ctx, cmd)
	return res, err
}

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1] // Get the type name without the package
}
