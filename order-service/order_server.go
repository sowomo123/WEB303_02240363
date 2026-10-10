package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "ecommerce-microservices/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// orderServer implements the Order Service gRPC API.
type orderServer struct {
	pb.UnimplementedOrderServiceServer

	db            *sql.DB
	productClient pb.ProductServiceClient
	breaker       *circuitBreaker
}

// getEnv returns an environment variable or a fallback value.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// getEnvInt reads an integer environment variable.
func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return result, nil
}

// openOrderDB connects to the dedicated Order PostgreSQL database.
func openOrderDB(ctx context.Context) (*sql.DB, error) {
	port, err := getEnvInt("ORDER_DB_PORT", 5432)
	if err != nil {
		return nil, err
	}

	host := getEnv("ORDER_DB_HOST", "localhost")
	user := getEnv("ORDER_DB_USER", "order_user")
	password := os.Getenv("ORDER_DB_PASSWORD")
	name := getEnv("ORDER_DB_NAME", "orders_db")

	if password == "" {
		return nil, errors.New("ORDER_DB_PASSWORD environment variable is required")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open order database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to order database: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			order_id VARCHAR(50) PRIMARY KEY,
			customer_id VARCHAR(50) NOT NULL,
			product_id VARCHAR(50) NOT NULL,
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			total_price DOUBLE PRECISION NOT NULL CHECK (total_price >= 0)
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create orders table: %w", err)
	}

	log.Println("Order Service connected to PostgreSQL")
	return db, nil
}

// connectProductService creates the gRPC client for Product Service.
func connectProductService(ctx context.Context) (*grpc.ClientConn, pb.ProductServiceClient, error) {
	address := getEnv("PRODUCT_SERVICE_ADDR", "localhost:50051")

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to Product Service: %w", err)
	}

	return conn, pb.NewProductServiceClient(conn), nil
}

// validateOrderInput checks required order fields.
func validateOrderInput(orderID, customerID, productID string, quantity int32) error {
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}
	if customerID == "" {
		return status.Error(codes.InvalidArgument, "customer_id is required")
	}
	if productID == "" {
		return status.Error(codes.InvalidArgument, "product_id is required")
	}
	if quantity <= 0 {
		return status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}

	return nil
}

// CreateOrder validates the product and stores a new order.
func (s *orderServer) CreateOrder(
	ctx context.Context,
	req *pb.CreateOrderRequest,
) (*pb.Order, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := validateOrderInput(
		req.GetOrderId(),
		req.GetCustomerId(),
		req.GetProductId(),
		req.GetQuantity(),
	); err != nil {
		return nil, err
	}

	product, err := getProduct(ctx, s.productClient, s.breaker, req.GetProductId())
	if err != nil {
		return nil, err
	}

	total := product.GetPrice() * float64(req.GetQuantity())

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO orders (order_id, customer_id, product_id, quantity, total_price)
		VALUES ($1, $2, $3, $4, $5)
	`,
		req.GetOrderId(),
		req.GetCustomerId(),
		req.GetProductId(),
		req.GetQuantity(),
		total,
	)
	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "could not create order: %v", err)
	}

	return &pb.Order{
		OrderId:    req.GetOrderId(),
		CustomerId: req.GetCustomerId(),
		ProductId:  req.GetProductId(),
		Quantity:   req.GetQuantity(),
		TotalPrice: total,
	}, nil
}

// GetOrder retrieves an order and verifies its product through gRPC.
func (s *orderServer) GetOrder(
	ctx context.Context,
	req *pb.GetOrderRequest,
) (*pb.Order, error) {
	if req == nil || req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	order := &pb.Order{}
	err := s.db.QueryRowContext(ctx, `
		SELECT order_id, customer_id, product_id, quantity, total_price
		FROM orders WHERE order_id = $1
	`, req.GetOrderId()).Scan(
		&order.OrderId,
		&order.CustomerId,
		&order.ProductId,
		&order.Quantity,
		&order.TotalPrice,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "retrieve order: %v", err)
	}

	_, err = getProduct(ctx, s.productClient, s.breaker, order.ProductId)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// UpdateOrder validates the product and updates an existing order.
func (s *orderServer) UpdateOrder(
	ctx context.Context,
	req *pb.UpdateOrderRequest,
) (*pb.Order, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := validateOrderInput(
		req.GetOrderId(),
		req.GetCustomerId(),
		req.GetProductId(),
		req.GetQuantity(),
	); err != nil {
		return nil, err
	}

	product, err := getProduct(ctx, s.productClient, s.breaker, req.GetProductId())
	if err != nil {
		return nil, err
	}

	total := product.GetPrice() * float64(req.GetQuantity())

	result, err := s.db.ExecContext(ctx, `
		UPDATE orders
		SET customer_id = $2, product_id = $3, quantity = $4, total_price = $5
		WHERE order_id = $1
	`,
		req.GetOrderId(),
		req.GetCustomerId(),
		req.GetProductId(),
		req.GetQuantity(),
		total,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update order: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check updated order: %v", err)
	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &pb.Order{
		OrderId:    req.GetOrderId(),
		CustomerId: req.GetCustomerId(),
		ProductId:  req.GetProductId(),
		Quantity:   req.GetQuantity(),
		TotalPrice: total,
	}, nil
}

// DeleteOrder removes an order from PostgreSQL.
func (s *orderServer) DeleteOrder(
	ctx context.Context,
	req *pb.DeleteOrderRequest,
) (*pb.DeleteOrderResponse, error) {
	if req == nil || req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	result, err := s.db.ExecContext(
		ctx,
		`DELETE FROM orders WHERE order_id = $1`,
		req.GetOrderId(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete order: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check deleted order: %v", err)
	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &pb.DeleteOrderResponse{
		Success: true,
		Message: "order deleted successfully",
	}, nil
}

// main starts the PostgreSQL-backed Order Service gRPC server.
func main() {
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := openOrderDB(dbCtx)
	dbCancel()
	if err != nil {
		log.Fatalf("Failed to connect to Order Database: %v", err)
	}
	defer db.Close()

	productCtx, productCancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, productClient, err := connectProductService(productCtx)
	productCancel()
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer conn.Close()

	port := getEnv("ORDER_SERVICE_PORT", "50052")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	breaker := &circuitBreaker{}

	pb.RegisterOrderServiceServer(server, &orderServer{
		db:            db,
		productClient: productClient,
		breaker:       breaker,
	})

	log.Printf("Order Service listening on port %s", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Order Service failed: %v", err)
	}
}
