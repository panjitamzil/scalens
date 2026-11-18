package observability

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var Buckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

type Metrics struct {
	Registry *prometheus.Registry
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
	Acquire  *prometheus.HistogramVec
	Query    *prometheus.HistogramVec
}

func NewMetrics(service string) *Metrics {
	r := prometheus.NewRegistry()
	labels := prometheus.Labels{"service": service}
	m := &Metrics{
		Registry: r,
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_server_requests_total", Help: "Completed HTTP requests.", ConstLabels: labels}, []string{"method", "route", "status_code", "outcome"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_server_request_duration_seconds", Help: "Server request duration.", Buckets: Buckets, ConstLabels: labels}, []string{"method", "route", "outcome"}),
		Acquire:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "db_client_acquire_duration_seconds", Help: "Time to acquire a connection.", Buckets: Buckets, ConstLabels: labels}, []string{"outcome"}),
		Query:    prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "db_client_query_duration_seconds", Help: "Database operation duration.", Buckets: Buckets, ConstLabels: labels}, []string{"operation", "outcome"}),
	}
	r.MustRegister(m.Requests, m.Duration, m.Acquire, m.Query)
	return m
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (m *Metrics) Wrap(route string, h http.HandlerFunc) http.Handler {
	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("http.route", route))
		rw := &responseWriter{ResponseWriter: w}
		h(rw, r)
		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}
		outcome := "failure"
		if status >= 200 && status < 300 {
			outcome = "success"
		}
		m.Requests.WithLabelValues(r.Method, route, strconv.Itoa(status), outcome).Inc()
		m.Duration.WithLabelValues(r.Method, route, outcome).Observe(time.Since(start).Seconds())
	}), route)
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

func Tracing(ctx context.Context, service string) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return func(context.Context) error { return nil }, nil
	}
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	res := resource.NewSchemaless(
		attribute.String("service.name", service),
		attribute.String("service.version", os.Getenv("APP_REVISION")),
		attribute.String("k8s.pod.name", os.Getenv("POD_NAME")),
		attribute.String("k8s.namespace.name", os.Getenv("POD_NAMESPACE")),
	)
	ratio := 1.0
	if raw := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); raw != "" {
		ratio, err = strconv.ParseFloat(raw, 64)
		if err != nil || ratio < 0 || ratio > 1 {
			return nil, fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be between 0 and 1")
		}
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res), sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))))
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
