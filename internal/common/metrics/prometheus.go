package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// impl MetricsClient
type PrometheusMetricsClient struct {
	registry *prometheus.Registry
}

var dynamicCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "dynamic-counter", Help: "count custom keys"}, []string{"key"},
)

type PrometheusMetricsClientConfig struct {
	Host        string
	ServiceName string
}

func NewPrometheusMetricsClient(cfg *PrometheusMetricsClientConfig) *PrometheusMetricsClient {
	cli := &PrometheusMetricsClient{}
	cli.initPrometheus(cfg)
	return cli
}

func (p *PrometheusMetricsClient) initPrometheus(cfg *PrometheusMetricsClientConfig) {
	p.registry = prometheus.NewRegistry()
	p.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	p.registry.Register(dynamicCounter) // regeister self-defined counter

	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": cfg.ServiceName}, p.registry)

	// export
	http.Handle("/metrics", promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}))
	go func() {
		logrus.Fatalf("Failed to start prometheus metrics endpoint err=%v", http.ListenAndServe(cfg.Host, nil))
	}()
}

func (p *PrometheusMetricsClient) Inc(key string, val int) {
	dynamicCounter.WithLabelValues(key).Add(float64(val))
}
