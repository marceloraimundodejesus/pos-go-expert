package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"service-b/internal/cep"
	"service-b/internal/core"
	httpx "service-b/internal/http"
	"service-b/internal/weather"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

func initTracer() func(context.Context) error {
	ctx := context.Background()
	conn, _ := grpc.DialContext(ctx, "otel-collector:4317", grpc.WithInsecure())
	exp, _ := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName("service-b"))),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp.Shutdown
}

func main() {
	shutdown := initTracer()
	defer shutdown(context.Background())

	http.DefaultTransport = otelhttp.NewTransport(http.DefaultTransport)
	http.DefaultClient.Timeout = 8 * time.Second

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	via := cep.NewViaCEP()
	wapi := weather.NewWeatherAPI()

	h := httpx.Handler{
		Svc: core.Service{
			ViaCEP:  via,
			Weather: wapi,
		},
	}

	log.Printf("listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, otelhttp.NewHandler(h, "service-b.server")); err != nil {
		log.Fatal(err)
	}
}
