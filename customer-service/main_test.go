package main

import (
	"context"
	"testing"

	pb "ecommerce-microservices/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetCustomer_ExistingCustomer(t *testing.T) {
	server := &customerServer{}

	customer, err := server.GetCustomer(
		context.Background(),
		&pb.GetCustomerRequest{CustomerId: "C001"},
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if customer.GetCustomerId() != "C001" {
		t.Errorf("expected C001, got %s", customer.GetCustomerId())
	}

	if customer.GetName() != "Sonam Wangmo" {
		t.Errorf("expected Sonam Wangmo, got %s", customer.GetName())
	}
}

func TestGetCustomer_NotFound(t *testing.T) {
	server := &customerServer{}

	customer, err := server.GetCustomer(
		context.Background(),
		&pb.GetCustomerRequest{CustomerId: "C999"},
	)

	if customer != nil {
		t.Errorf("expected nil customer, got: %v", customer)
	}

	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got: %v", status.Code(err))
	}
}
