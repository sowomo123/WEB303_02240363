package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./order-service <product-id>")
		fmt.Println("Example: go run ./order-service P001")
		return
	}

	productID := os.Args[1]

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("could not connect to Product Service: %v", err)
	}

	defer conn.Close()

	client := pb.NewProductServiceClient(conn)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	product, err := client.GetProduct(
		ctx,
		&pb.GetProductRequest{
			ProductId: productID,
		},
	)

	if err != nil {
		log.Fatalf("GetProduct failed: %v", err)
	}

	fmt.Println("========== PRODUCT INFORMATION ==========")
	fmt.Printf("Product ID: %s\n", product.GetProductId())
	fmt.Printf("Name: %s\n", product.GetName())
	fmt.Printf("Price: %.2f\n", product.GetPrice())
	fmt.Println("=========================================")
}