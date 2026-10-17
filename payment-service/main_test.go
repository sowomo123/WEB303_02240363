
package main

import (
	"context"
	"strings"
	"testing"

	pb "ecommerce-microservices/proto"
)

func TestProcessPayment_Success(t *testing.T) {
	server := &paymentServer{}

	resp, err := server.ProcessPayment(context.Background(),
		&pb.PaymentRequest{
			CustomerId:   "C001",
			Amount:       1000,
			PaymentMethod: "CARD",
		})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !resp.GetSuccess() {
		t.Error("expected payment to succeed")
	}
	if resp.GetMessage() != "Payment processed successfully" {
		t.Errorf("unexpected message: %s", resp.GetMessage())
	}
}

func TestProcessPayment_TransactionIDFormat(t *testing.T) {
	server := &paymentServer{}

	resp, err := server.ProcessPayment(context.Background(),
		&pb.PaymentRequest{
			CustomerId:   "C001",
			Amount:       500,
			PaymentMethod: "CASH",
		})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.HasPrefix(resp.GetTransactionId(), "TXN-") {
		t.Errorf("expected transaction ID to start with TXN-, got %s",
			resp.GetTransactionId())
	}
	if len(resp.GetTransactionId()) <= len("TXN-") {
		t.Errorf("expected transaction ID to contain a number, got %s",
			resp.GetTransactionId())
	}
}

func TestProcessPayment_DifferentAmounts(t *testing.T) {
	server := &paymentServer{}

	amounts := []float64{1, 250, 9999.99}

	for _, amount := range amounts {
		resp, err := server.ProcessPayment(context.Background(),
			&pb.PaymentRequest{
				CustomerId:    "C002",
				Amount:        amount,
				PaymentMethod: "CARD",
			})

		if err != nil {
			t.Errorf("amount %.2f: expected no error, got %v", amount, err)
			continue
		}
		if !resp.GetSuccess() {
			t.Errorf("amount %.2f: expected payment success", amount)
		}
	}
}

func TestProcessPayment_DifferentPaymentMethods(t *testing.T) {
	server := &paymentServer{}

	methods := []string{"CARD", "CASH", "ONLINE"}

	for _, method := range methods {
		resp, err := server.ProcessPayment(context.Background(),
			&pb.PaymentRequest{
				CustomerId:    "C003",
				Amount:        100,
				PaymentMethod: method,
			})

		if err != nil {
			t.Errorf("method %s: expected no error, got %v", method, err)
			continue
		}
		if !resp.GetSuccess() {
			t.Errorf("method %s: expected payment success", method)
		}
	}
}
