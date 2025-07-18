package metrics

type MetricsClient interface {
	Inc(key string, val int)
}

type NoMetrics struct{}

func (m NoMetrics) Inc(_ string, _ int) {} // do nothing
