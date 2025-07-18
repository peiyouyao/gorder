package decorator

import (
	"context"

	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/sirupsen/logrus"
)

type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

func ApplyCommandDecorators[C, R any](
	handler CommandHandler[C, R],
	logger *logrus.Entry,
	metrics metrics.MetricsClient,
) CommandHandler[C, R] {
	return commandLoggingDecorator[C, R]{
		logger: logger,
		handler: commandMetricsDecorator[C, R]{
			handler: handler,
			metrics: metrics,
		},
	}
}
