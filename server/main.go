package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	pb "github.com/HardDie/grpc_with_tracing_example/pkg/server"
)

const (
	grpcPort = ":9000"
)

func main() {
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
