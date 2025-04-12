package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	lambdadetector "go.opentelemetry.io/contrib/detectors/aws/lambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("no IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("non 200 Response found")
)

var tracer trace.Tracer

func instrumentAWSClients(ctx context.Context) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v\n", err)
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)
}

func setupOtlpTracer(ctx context.Context, res *resource.Resource) *sdktrace.TracerProvider {
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("failed to initialize optl trace exporter over grcp %v\n", err)
		return nil
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(traceExporter),
		sdktrace.WithResource(
			res,
		),
	)

	return traceProvider
}

func setupOtlpMeter(ctx context.Context, res *resource.Resource) *sdkmetric.MeterProvider {
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("failed to initialize optl metric exporter over grcp %v\n", err)
		return nil
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter),
		),
	)

	return meterProvider
}

func setupOtlpLogger(ctx context.Context, res *resource.Resource) *sdklog.LoggerProvider {
	logExporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint("localhost:4317"),
		otlploggrpc.WithInsecure(),
	)

	if err != nil {
		log.Printf("failed to initialize optl log exporter over grcp %v\n", err)
		return nil
	}

	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(logExporter),
		),
	)

	return logProvider
}

func setupOtlpSdk(ctx context.Context) {
	detector := lambdadetector.NewResourceDetector()
	resource, err := detector.Detect(ctx)
	if err != nil {
		log.Fatalf("failed to detect lambda resources: %v\n", err)
		return
	}

	optlPropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(optlPropagator)

	otel.SetTracerProvider(
		setupOtlpTracer(
			ctx,
			resource,
		),
	)
	otel.SetMeterProvider(
		setupOtlpMeter(
			ctx,
			resource,
		),
	)
	global.SetLoggerProvider(
		setupOtlpLogger(
			ctx,
			resource,
		),
	)

}

func getIpAddress(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "getIpAddressSpan")
	defer span.End()

	resp, err := otelhttp.Get(ctx, DefaultHTTPGetAddress)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", ErrNon200Response
	}

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if len(ip) == 0 {
		return "", ErrNoIP
	}

	return string(ip), nil
}

func lambdaHandler(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	ctx, span := tracer.Start(ctx, "lambdaHandlerSpan")
	defer span.End()

	ip, err := getIpAddress(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", string(ip)),
		StatusCode: 200,
	}, nil
}

func main() {
	ctx := context.Background()

	// setup otlp sdk
	setupOtlpSdk(ctx)

	// instrument aws clients
	instrumentAWSClients(ctx)

	// get tracer
	traceProvider := otel.GetTracerProvider()
	defer func() {
		if sdkTracerProvider, ok := traceProvider.(*sdktrace.TracerProvider); ok {
			_ = sdkTracerProvider.Shutdown(ctx)
		}
	}()

	tracer = traceProvider.Tracer("dynova.io/HelloWorld")

	lambda.Start(
		otellambda.InstrumentHandler(
			lambdaHandler,
			otellambda.WithTracerProvider(otel.GetTracerProvider()),
			otellambda.WithPropagator(otel.GetTextMapPropagator()),
		),
	)
}
