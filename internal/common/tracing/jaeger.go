package tracing

import (
	"context"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// 全局 Tracer 对象
var tracer = otel.Tracer("default_tracer")

// 初始化 Jaeger 链路追踪 Provider
func InitJaegerProvider(jaegerURL, serviceName string) (func(ctx context.Context) error, error) {
	if jaegerURL == "" {
		panic("empty jaeger url") // 必须配置 Jaeger URL
	}

	tracer = otel.Tracer(serviceName) // 设置当前服务的 Tracer 名字

	// 创建 Jaeger Exporter（负责发送 Trace 数据到 Jaeger）
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(jaegerURL),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建 TracerProvider（管理所有 Tracer）
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp), // 批量发送 Trace
		sdktrace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName))), // 标记服务名
	)

	// 注册为全局 Provider
	otel.SetTracerProvider(tp)

	// 设置上下文传播器（支持 W3C TraceContext）
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// 兼容 B3 格式（Zipkin/Jaeger）
	b3Propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}, b3Propagator)
	otel.SetTextMapPropagator(p)

	// 返回关闭函数，用于应用退出时 flush 数据
	return tp.Shutdown, nil
}

// 创建新 Span 的封装
func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

// 获取当前 TraceID
func TraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	return spanCtx.TraceID().String()
}
