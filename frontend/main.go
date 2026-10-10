package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	productClient pb.ProductServiceClient
}

func (s *Server) getProduct(w http.ResponseWriter, r *http.Request) {

	productID := strings.TrimPrefix(r.URL.Path, "/product/")
	productID = strings.TrimSpace(productID)

	w.Header().Set("Content-Type", "application/json")

	// Empty ID
	if productID == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{
			"error": "Please enter a Product ID.",
		})

		return
	}

	// Set timeout for gRPC request
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)

	defer cancel()

	// Call Product Service through gRPC
	product, err := s.productClient.GetProduct(
		ctx,
		&pb.GetProductRequest{
			ProductId: productID,
		},
	)

	if err != nil {

		code := status.Code(err)

		switch code {

		case codes.NotFound:

			w.WriteHeader(http.StatusNotFound)

			json.NewEncoder(w).Encode(map[string]string{
				"error": "Product not found.",
			})

		case codes.DeadlineExceeded:

			w.WriteHeader(http.StatusGatewayTimeout)

			json.NewEncoder(w).Encode(map[string]string{
				"error": "Product service request timed out.",
			})

		case codes.Unavailable:

			w.WriteHeader(http.StatusServiceUnavailable)

			json.NewEncoder(w).Encode(map[string]string{
				"error": "Product service is unavailable.",
			})

		default:

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]string{
				"error": "An internal error occurred.",
			})
		}

		return
	}

	// Successful response
	response := map[string]interface{}{
		"product_id": product.GetProductId(),
		"name":       product.GetName(),
		"price":      product.GetPrice(),
	}

	json.NewEncoder(w).Encode(response)
}

func main() {

	// Connect to Product Service
	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}

	defer conn.Close()

	productClient := pb.NewProductServiceClient(conn)

	server := &Server{
		productClient: productClient,
	}

	// Serve static files
	http.Handle(
		"/",
		http.FileServer(http.Dir(".")),
	)

	// Product API
	http.HandleFunc(
		"/product/",
		server.getProduct,
	)

	fmt.Println("Frontend server running on http://localhost:8081")

	log.Fatal(
		http.ListenAndServe(":8081", nil),
	)
}