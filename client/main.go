package main

import (
	"context"
	"fmt"
	"log"
	"net"

	jaegerExporter "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	pb "github.com/HardDie/grpc_with_tracing_example/pkg/client"
	"github.com/HardDie/grpc_with_tracing_example/pkg/server"
)

const (
	grpcPort = ":9001"
)

var (
	// Store a global trace provider variable to clear it before closing
	tracer *tracesdk.TracerProvider
)

func NewTracer(url, name string) error {
	// Create the Jaeger exporter
	exp, err := jaegerExporter.New(jaegerExporter.WithCollectorEndpoint(jaegerExporter.WithEndpoint(url)))
	if err != nil {
		return err
	}
	tracer = tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
		)),
	)
	return nil
}

func main() {
	err := NewTracer("http://localhost:14268/api/traces", "client")
	if err != nil {
		log.Fatal(err)
	}
	defer tracer.Shutdown(context.Background())

	// Create a TCP connection
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatal(err)
	}

	// Create the GRPC server
	grpcServer := grpc.NewServer()

	// Allows us to use a 'list' call to list all available APIs
	reflection.Register(grpcServer)

	// We register an object that should implement all the described APIs
	pb.RegisterClientServer(grpcServer, &ClientServeObject{})

	// Serving the GRPC server on a created TCP socket
	log.Println("GRPC server listening on " + grpcPort)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

// ClientServeObject Describe the structure that should implement the interface described in the proto file
type ClientServeObject struct {
	pb.UnimplementedClientServer
}

// Test Implement a only endpoint
func (s *ClientServeObject) Test(ctx context.Context, _ *pb.TestRequest) (*pb.TestResponse, error) {
	ctx, span := tracer.Tracer("client").Start(ctx, "Test")
	defer span.End()

	traceId := fmt.Sprintf("%s", span.SpanContext().TraceID())
	ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceId)

	// Create a connection to the server
	conn, err := grpc.DialContext(ctx, "localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Create a client object with connection
	serv := server.NewServerClient(conn)

	// Calling a method on the server side
	resp, err := serv.Test(ctx, &server.TestRequest{})
	if err != nil {
		return nil, err
	}

	// Forwarding the response from the server to the client
	return &pb.TestResponse{
		Message: resp.GetMessage(),
	}, nil
}
