package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	pb "ecommerce-microservices/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	requestTimeout = 3 * time.Second
	maxRetries     = 2
	failureLimit   = 3
	openDuration   = 5 * time.Second
)

type circuitBreaker struct {
	mu       sync.Mutex
	failures int
	openedAt time.Time
	open     bool
}

func (b *circuitBreaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.open {
		return true
	}

	if time.Since(b.openedAt) >= openDuration {
		b.open = false
		b.failures = 0
		log.Println("circuit breaker half-open: allowing a trial request")
		return true
	}

	return false
}

func (b *circuitBreaker) recordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures = 0
	b.open = false
}

func (b *circuitBreaker) recordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++

	if b.failures >= failureLimit && !b.open {
		b.open = true
		b.openedAt = time.Now()

		log.Printf(
			"circuit breaker opened after %d failures",
			b.failures,
		)
	}
}

func getProduct(
	ctx context.Context,
	client pb.ProductServiceClient,
	breaker *circuitBreaker,
	productID string,
) (*pb.Product, error) {

	if !breaker.allow() {
		return nil, status.Error(
			codes.Unavailable,
			"circuit breaker is open",
		)
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {

		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}

		attemptCtx, cancel := context.WithTimeout(
			ctx,
			requestTimeout,
		)

		log.Printf(
			"GetProduct attempt %d/%d",
			attempt+1,
			maxRetries+1,
		)

		product, err := client.GetProduct(
			attemptCtx,
			&pb.GetProductRequest{
				ProductId: productID,
			},
		)

		cancel()

		if err == nil {
			breaker.recordSuccess()
			return product, nil
		}

		lastErr = err

		code := status.Code(err)

		log.Printf(
			"attempt %d failed with %s: %v",
			attempt+1,
			code,
			err,
		)

		if code != codes.Unavailable &&
			code != codes.DeadlineExceeded &&
			code != codes.ResourceExhausted {
			break
		}
	}

	breaker.recordFailure()

	return nil, lastErr
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <product-id>")
		fmt.Println("Example: go run . P001")
		return
	}

	productID := os.Args[1]

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"could not connect to Product Service: %v",
			err,
		)
	}

	defer conn.Close()

	client := pb.NewProductServiceClient(conn)

	breaker := &circuitBreaker{}

	// Normal mode
	if os.Getenv("CIRCUIT_DEMO") != "true" {

		ctx, cancel := context.WithTimeout(
			context.Background(),
			requestTimeout,
		)
		defer cancel()

		product, err := getProduct(
			ctx,
			client,
			breaker,
			productID,
		)

		if err != nil {
			log.Printf(
				"GetProduct failed: %v",
				err,
			)
			return
		}

		fmt.Println(
			"========== PRODUCT INFORMATION ==========",
		)
		fmt.Printf(
			"Product ID: %s\n",
			product.GetProductId(),
		)
		fmt.Printf(
			"Name: %s\n",
			product.GetName(),
		)
		fmt.Printf(
			"Price: %.2f\n",
			product.GetPrice(),
		)
		fmt.Println(
			"=========================================",
		)

		return
	}

	// Circuit Breaker demonstration
	fmt.Println(
		"========== CIRCUIT BREAKER DEMO ==========",
	)

	for i := 1; i <= 5; i++ {

		fmt.Printf("\n--- Request %d ---\n", i)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			requestTimeout,
		)

		product, err := getProduct(
			ctx,
			client,
			breaker,
			productID,
		)

		cancel()

		if err != nil {
			fmt.Printf(
				"Request %d failed: %v\n",
				i,
				err,
			)
		} else {
			fmt.Println("Request successful")
			fmt.Printf(
				"Product: %s\n",
				product.GetName(),
			)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println(
		"\n========== WAITING FOR CIRCUIT RECOVERY ==========",
	)
	fmt.Println("Waiting 6 seconds...")

	time.Sleep(6 * time.Second)

	fmt.Println(
		"\n--- Trial Request After Recovery ---",
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		requestTimeout,
	)

	product, err := getProduct(
		ctx,
		client,
		breaker,
		productID,
	)

	cancel()

	if err != nil {
		fmt.Printf(
			"Trial request failed: %v\n",
			err,
		)
	} else {
		fmt.Println("Trial request successful")
		fmt.Printf(
			"Product: %s\n",
			product.GetName(),
		)
	}

	fmt.Println(
		"==============================================",
	)
}