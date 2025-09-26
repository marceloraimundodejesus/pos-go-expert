package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

type reqBody struct {
	CEP string `json:"cep"`
}

var reCEP = regexp.MustCompile(`^\d{8}$`)

func initTracer(service string) func(context.Context) error {
	ctx := context.Background()
	conn, _ := grpc.DialContext(ctx, "otel-collector:4317", grpc.WithInsecure())
	exp, _ := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(service))),
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
	shutdown := initTracer("service-a")
	defer shutdown(context.Background())

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   8 * time.Second,
	}

	mux := http.NewServeMux()

	mux.Handle("/echo", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		raw, _ := io.ReadAll(r.Body)
		var in reqBody
		_ = json.Unmarshal(raw, &in)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		out := map[string]any{
			"raw":       string(raw),
			"json_cep":  in.CEP,
			"len_bytes": len(raw),
			"ctype":     r.Header.Get("Content-Type"),
		}
		b, _ := json.Marshal(out)
		w.Write(b)
	}), "service-a.echo"))

	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		raw, err := io.ReadAll(r.Body)
		if err != nil || len(raw) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		}

		var in reqBody
		if err := json.Unmarshal(raw, &in); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		}

		cep := strings.TrimSpace(in.CEP)
		if !reCEP.MatchString(cep) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("invalid zipcode"))
			return
		}

		base := os.Getenv("SERVICE_B_URL")
		if base == "" {
			base = "http://localhost:8081"
		}
		url := fmt.Sprintf("%s/weather/%s", base, cep)

		req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("service b unavailable"))
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		ct := resp.Header.Get("Content-Type")
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		} else {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(body)

	}), "service-a.root"))

	http.ListenAndServe(":8080", mux)
}
