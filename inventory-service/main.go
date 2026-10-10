package main

import (
    "context"
    "log"
    "net"

    pb "ecommerce-microservices/proto"

    "google.golang.org/grpc"
)

type inventoryServer struct {
	pb.UnimplementedInventoryServiceServer
}

func (s *inventoryServer) CheckStock(
	ctx context.Context,
	req *pb.CheckStockRequest,
) (*pb.StockResponse, error) {

	log.Printf(
		"Checking stock for product %s, requested quantity: %d",
		req.GetProductId(),
		req.GetQuantity(),
	)

	// Demo stock quantity
	availableQuantity := int32(10)

	if req.GetQuantity() <= availableQuantity {
		return &pb.StockResponse{
			ProductId:         req.GetProductId(),
			AvailableQuantity: availableQuantity - req.GetQuantity(),
			Available:         true,
		}, nil
	}

	return &pb.StockResponse{
		ProductId:         req.GetProductId(),
		AvailableQuantity: availableQuantity,
		Available:         false,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterInventoryServiceServer(
		grpcServer,
		&inventoryServer{},
	)

	log.Println("Inventory service running on :50054")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
