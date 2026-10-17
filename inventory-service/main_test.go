
package main

import (
	"context"
	"testing"

	pb "ecommerce-microservices/proto"
)

func TestCheckStock_Available(t *testing.T) {
	server := &inventoryServer{}

	resp, err := server.CheckStock(context.Background(),
		&pb.CheckStockRequest{
			ProductId: "P001",
			Quantity:  3,
		})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !resp.GetAvailable() {
		t.Error("expected stock to be available")
	}
	if resp.GetAvailableQuantity() != 7 {
		t.Errorf("expected remaining quantity 7, got %d",
			resp.GetAvailableQuantity())
	}
	if resp.GetProductId() != "P001" {
		t.Errorf("expected product ID P001, got %s",
			resp.GetProductId())
	}
}

func TestCheckStock_ExactAvailableQuantity(t *testing.T) {
	server := &inventoryServer{}

	resp, err := server.CheckStock(context.Background(),
		&pb.CheckStockRequest{
			ProductId: "P002",
			Quantity:  10,
		})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !resp.GetAvailable() {
		t.Error("expected stock to be available for quantity 10")
	}
	if resp.GetAvailableQuantity() != 0 {
		t.Errorf("expected remaining quantity 0, got %d",
			resp.GetAvailableQuantity())
	}
}

func TestCheckStock_InsufficientStock(t *testing.T) {
	server := &inventoryServer{}

	resp, err := server.CheckStock(context.Background(),
		&pb.CheckStockRequest{
			ProductId: "P003",
			Quantity:  15,
		})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.GetAvailable() {
		t.Error("expected stock to be unavailable")
	}
	if resp.GetAvailableQuantity() != 10 {
		t.Errorf("expected available quantity 10, got %d",
			resp.GetAvailableQuantity())
	}
}

func TestCheckStock_ZeroQuantity(t *testing.T) {
	server := &inventoryServer{}

	resp, err := server.CheckStock(context.Background(),
		&pb.CheckStockRequest{
			ProductId: "P004",
			Quantity:  0,
		})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !resp.GetAvailable() {
		t.Error("expected stock to be available for quantity 0")
	}
	if resp.GetAvailableQuantity() != 10 {
		t.Errorf("expected available quantity 10, got %d",
			resp.GetAvailableQuantity())
	}
}


