package main

import (
	"context"
	"log"
	"net"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type customerServer struct {
	pb.UnimplementedCustomerServiceServer
}

var customers = map[string]*pb.Customer{
	"C001": {
		CustomerId: "C001",
		Name:       "Sonam Wangmo",
		Email:      "sonam@example.com",
	},
	"C002": {
		CustomerId: "C002",
		Name:       "Pema Dorji",
		Email:      "pema@example.com",
	},
	"C003": {
		CustomerId: "C003",
		Name:       "Karma Tshering",
		Email:      "karma@example.com",
	},
}

func (s *customerServer) GetCustomer(
	ctx context.Context,
	req *pb.GetCustomerRequest,
) (*pb.Customer, error) {

	customer, exists := customers[req.GetCustomerId()]

	if !exists {
		return nil, status.Errorf(
			codes.NotFound,
			"customer %s not found",
			req.GetCustomerId(),
		)
	}

	return customer, nil
}

func main() {

	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterCustomerServiceServer(
		grpcServer,
		&customerServer{},
	)

	// Enable grpcurl
	reflection.Register(grpcServer)

	log.Println("Customer Service running on port 50053")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}