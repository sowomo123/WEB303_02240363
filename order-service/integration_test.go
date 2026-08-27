package main

import (
	"context"
	"testing"
	"time"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// connectToProductService creates a real gRPC connection
// to the Product Service.
func connectToProductService(t *testing.T) (
	pb.ProductServiceClient,
	*grpc.ClientConn,
) {

	t.Helper()

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		t.Fatalf(
			"failed to create gRPC connection: %v",
			err,
		)
	}

	client := pb.NewProductServiceClient(conn)

	return client, conn
}

// ============================================================
// TEST C1: Product Service Available
// ============================================================

func TestIntegration_ProductService_Available(t *testing.T) {

	client, conn := connectToProductService(t)
	defer conn.Close()

	breaker := &circuitBreaker{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	product, err := getProduct(
		ctx,
		client,
		breaker,
		"P001",
	)

	if err != nil {
		t.Fatalf(
			"expected successful response, got error: %v",
			err,
		)
	}

	if product == nil {
		t.Fatal("expected product, got nil")
	}

	if product.GetProductId() != "P001" {
		t.Errorf(
			"expected P001, got %s",
			product.GetProductId(),
		)
	}

	if product.GetName() != "Laptop" {
		t.Errorf(
			"expected Laptop, got %s",
			product.GetName(),
		)
	}

	if product.GetPrice() != 75000 {
		t.Errorf(
			"expected price 75000, got %v",
			product.GetPrice(),
		)
	}
}

// ============================================================
// TEST C2: Different Valid Identifiers
// ============================================================

func TestIntegration_DifferentProducts(t *testing.T) {

	client, conn := connectToProductService(t)
	defer conn.Close()

	testCases := []struct {
		id    string
		name  string
		price float64
	}{
		{
			id:    "P001",
			name:  "Laptop",
			price: 75000,
		},
		{
			id:    "P002",
			name:  "Mechanical Keyboard",
			price: 4500,
		},
		{
			id:    "P003",
			name:  "Wireless Mouse",
			price: 1800,
		},
		{
			id:    "P004",
			name:  "Monitor",
			price: 25000,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.id, func(t *testing.T) {

			breaker := &circuitBreaker{}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			product, err := getProduct(
				ctx,
				client,
				breaker,
				tc.id,
			)

			if err != nil {
				t.Fatalf(
					"expected no error, got: %v",
					err,
				)
			}

			if product.GetProductId() != tc.id {
				t.Errorf(
					"expected ID %s, got %s",
					tc.id,
					product.GetProductId(),
				)
			}

			if product.GetName() != tc.name {
				t.Errorf(
					"expected name %s, got %s",
					tc.name,
					product.GetName(),
				)
			}

			if product.GetPrice() != tc.price {
				t.Errorf(
					"expected price %v, got %v",
					tc.price,
					product.GetPrice(),
				)
			}
		})
	}
}

// ============================================================
// TEST C3: Invalid Product Identifier
// ============================================================

func TestIntegration_ProductNotFound(t *testing.T) {

	client, conn := connectToProductService(t)
	defer conn.Close()

	breaker := &circuitBreaker{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	product, err := getProduct(
		ctx,
		client,
		breaker,
		"P999",
	)

	if product != nil {
		t.Errorf(
			"expected nil product, got %v",
			product,
		)
	}

	if err == nil {
		t.Fatal("expected NotFound error, got nil")
	}

	if status.Code(err) != codes.NotFound {
		t.Errorf(
			"expected NotFound, got %v",
			status.Code(err),
		)
	}
}

// ============================================================
// TEST C4: Product Service Unavailable
// ============================================================

func TestIntegration_ProductServiceUnavailable(t *testing.T) {

	client, conn := connectToProductService(t)
	defer conn.Close()

	breaker := &circuitBreaker{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	_, err := getProduct(
		ctx,
		client,
		breaker,
		"P001",
	)

	if err == nil {
		t.Fatal(
			"expected Product Service unavailable error, got nil",
		)
	}

	t.Logf(
		"Product Service unavailable handled correctly: %v",
		err,
	)
}

// ============================================================
// TEST C5: Product Service Returns Error
// ============================================================

func TestIntegration_ProductServiceError(t *testing.T) {

	client, conn := connectToProductService(t)
	defer conn.Close()

	breaker := &circuitBreaker{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	_, err := getProduct(
		ctx,
		client,
		breaker,
		"P001",
	)

	if err == nil {
		t.Fatal(
			"expected Product Service error, got nil",
		)
	}

	t.Logf(
		"Product Service error handled correctly: %v",
		err,
	)
}

// ============================================================
// TEST C6: Product Service Timeout
// ============================================================

func TestIntegration_ProductServiceTimeout(t *testing.T) {

	client, conn := connectToProductService(t)
	defer conn.Close()

	breaker := &circuitBreaker{}

	// The Product Service should be configured with
	// PRODUCT_DELAY=5s for this test.
	//
	// Order Service has a 3-second request timeout.

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	start := time.Now()

	_, err := getProduct(
		ctx,
		client,
		breaker,
		"P001",
	)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal(
			"expected timeout error, got nil",
		)
	}

	if status.Code(err) != codes.DeadlineExceeded {
		t.Errorf(
			"expected DeadlineExceeded, got %v",
			status.Code(err),
		)
	}

	t.Logf(
		"Timeout handled correctly after %v",
		elapsed,
	)
}
