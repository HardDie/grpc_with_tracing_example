package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	pb "github.com/HardDie/grpc_with_tracing_example/pkg/server"

	jaegerExporter "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	grpcPort = ":9000"
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
	err := NewTracer("http://localhost:14268/api/traces", "server")
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
	pb.RegisterServerServer(grpcServer, &ServerServeObject{})

	// Serving the GRPC server on a created TCP socket
	log.Println("GRPC server listening on " + grpcPort)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

// ServerServeObject Describe the structure that should implement the interface described in the proto file
type ServerServeObject struct {
	pb.UnimplementedServerServer
}

// Test Implement a only endpoint
func (s *ServerServeObject) Test(ctx context.Context, _ *pb.TestRequest) (*pb.TestResponse, error) {
	ctx, span := tracer.Tracer("server").Start(ctx, "Test")
	defer span.End()

	// Extract username from incoming context
	var username string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		val := md["username"]
		if len(val) > 0 {
			username = val[0]
		}
	}

	return &pb.TestResponse{
		Message: "Server response: " + username,
	}, nil
}
