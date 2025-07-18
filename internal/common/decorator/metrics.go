package decorator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/peiyouyao/gorder/common/metrics"
)

// impl QueryHandler interface
type queryMetricsDecorator[Q, R any] struct {
	handler QueryHandler[Q, R]
	metrics metrics.MetricsClient
}

func (d queryMetricsDecorator[Q, R]) Handle(ctx context.Context, query Q) (result R, err error) {
	start := time.Now()
	actionName := strings.ToLower(actionName(query))
	defer func() {
		end := time.Since(start)
		d.metrics.Inc(fmt.Sprintf("querys.%s.duration", actionName), int(end.Seconds()))
		if err == nil {
			d.metrics.Inc(fmt.Sprintf("querys.%s.success", actionName), 1)
		} else {
			d.metrics.Inc(fmt.Sprintf("querys.%s.failure", actionName), 1)
		}
	}()
	return d.handler.Handle(ctx, query)
}

// impl CommandHandler interface
type commandMetricsDecorator[C, R any] struct {
	handler QueryHandler[C, R]
	metrics metrics.MetricsClient
}

// impl QueryHandler interface
func (d commandMetricsDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	start := time.Now()
	actionName := strings.ToLower(actionName(cmd))
	defer func() {
		end := time.Since(start)
		d.metrics.Inc(fmt.Sprintf("commands.%s.duration", actionName), int(end.Seconds()))
		if err == nil {
			d.metrics.Inc(fmt.Sprintf("commands.%s.success", actionName), 1)
		} else {
			d.metrics.Inc(fmt.Sprintf("commands.%s.failure", actionName), 1)
		}
	}()
	return d.handler.Handle(ctx, cmd)
}
