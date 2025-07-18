package decorator

import (
	"context"

	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/sirupsen/logrus"
)

// QueryHandler defines a generic type that receives a Query Q, and returns a Result R.
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

func ApplyQueryDecorators[Q, R any](
	handler QueryHandler[Q, R],
	logger *logrus.Entry,
	metrics metrics.MetricsClient,
) QueryHandler[Q, R] {
	return queryLoggingDecorator[Q, R]{
		logger: logger,
		handler: queryMetricsDecorator[Q, R]{
			handler: handler,
			metrics: metrics,
		},
	}
}
