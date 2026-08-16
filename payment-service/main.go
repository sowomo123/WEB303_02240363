package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
)

type paymentServer struct {
	pb.UnimplementedPaymentServiceServer
}

func (s *paymentServer) ProcessPayment(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {

	log.Printf(
		"Processing payment for customer %s, amount: %.2f, method: %s",
		req.GetCustomerId(),
		req.GetAmount(),
		req.GetPaymentMethod(),
	)

	// Generate a simple transaction ID
	transactionID := fmt.Sprintf(
		"TXN-%d",
		time.Now().Unix(),
	)

	return &pb.PaymentResponse{
		Success:       true,
		TransactionId: transactionID,
		Message:       "Payment processed successfully",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterPaymentServiceServer(
		grpcServer,
		&paymentServer{},
	)

	log.Println("Payment service running on :50055")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}