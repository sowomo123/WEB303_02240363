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
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type productServer struct {
	pb.UnimplementedProductServiceServer
	db *sql.DB
}

// Retain the sample data for unit tests that instantiate productServer without a DB.
var products = map[string]*pb.Product{
	"P001": {ProductId: "P001", Name: "Laptop", Price: 75000},
	"P002": {ProductId: "P002", Name: "Mechanical Keyboard", Price: 4500},
	"P003": {ProductId: "P003", Name: "Wireless Mouse", Price: 1800},
	"P004": {ProductId: "P004", Name: "Monitor", Price: 25000},
}

func (s *productServer) getProduct(ctx context.Context, id string) (*pb.Product, error) {
	if s.db == nil {
		p, ok := products[id]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "product %s not found", id)
		}
		copy := *p
		return &copy, nil
	}

	p := &pb.Product{}
	err := s.db.QueryRowContext(ctx,
		`SELECT product_id, name, price FROM products WHERE product_id = $1`, id,
	).Scan(&p.ProductId, &p.Name, &p.Price)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "product %s not found", id)
	}
	if err != nil {
		log.Printf("database lookup failed: %v", err)
		return nil, status.Error(codes.Unavailable, "product database is unavailable")
	}
	return p, nil
}

func (s *productServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "product name is required")
	}
	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product price must be greater than zero")
	}

	p := &pb.Product{ProductId: req.GetProductId(), Name: req.GetName(), Price: req.GetPrice()}
	if s.db == nil {
		if _, exists := products[p.ProductId]; exists {
			return nil, status.Errorf(codes.AlreadyExists, "product %s already exists", p.ProductId)
		}
		products[p.ProductId] = p
		return p, nil
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO products (product_id, name, price) VALUES ($1, $2, $3)`,
		p.ProductId, p.Name, p.Price,
	)
	if err != nil {
		var pgErr interface{ SQLState() string }
		if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
			return nil, status.Errorf(codes.AlreadyExists, "product %s already exists", p.ProductId)
		}
		log.Printf("database insert failed: %v", err)
		return nil, status.Error(codes.Unavailable, "could not save product")
	}
	return p, nil
}

func (s *productServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "product name is required")
	}
	if req.GetPrice() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product price must be greater than zero")
	}

	if s.db == nil {
		p, ok := products[req.GetProductId()]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "product %s not found", req.GetProductId())
		}
		p.Name, p.Price = req.GetName(), req.GetPrice()
		return p, nil
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE products SET name = $1, price = $2 WHERE product_id = $3`,
		req.GetName(), req.GetPrice(), req.GetProductId(),
	)
	if err != nil {
		log.Printf("database update failed: %v", err)
		return nil, status.Error(codes.Unavailable, "could not update product")
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not verify product update")
	}
	if n == 0 {
		return nil, status.Errorf(codes.NotFound, "product %s not found", req.GetProductId())
	}
	return s.getProduct(ctx, req.GetProductId())
}

func (s *productServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if s.db == nil {
		if _, ok := products[req.GetProductId()]; !ok {
			return nil, status.Errorf(codes.NotFound, "product %s not found", req.GetProductId())
		}
		delete(products, req.GetProductId())
		return &pb.DeleteProductResponse{Success: true, Message: "product deleted successfully"}, nil
	}

	result, err := s.db.ExecContext(ctx, `DELETE FROM products WHERE product_id = $1`, req.GetProductId())
	if err != nil {
		log.Printf("database delete failed: %v", err)
		return nil, status.Error(codes.Unavailable, "could not delete product")
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not verify product deletion")
	}
	if n == 0 {
		return nil, status.Errorf(codes.NotFound, "product %s not found", req.GetProductId())
	}
	return &pb.DeleteProductResponse{Success: true, Message: "product deleted successfully"}, nil
}

func (s *productServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	if delay, err := time.ParseDuration(os.Getenv("PRODUCT_DELAY")); err == nil && delay > 0 {
		log.Printf("delaying GetProduct by %s", delay)
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, "product lookup timed out")
		}
	}
	if fail, _ := strconv.ParseBool(os.Getenv("PRODUCT_FAIL")); fail {
		return nil, status.Error(codes.Unavailable, "product service is temporarily unavailable")
	}
	return s.getProduct(ctx, req.GetProductId())
}

func databaseConfig() (string, error) {
	host := os.Getenv("PRODUCT_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PRODUCT_DB_PORT")
	if port == "" {
		port = "5433"
	}
	user := os.Getenv("PRODUCT_DB_USER")
	if user == "" {
		user = "product_user"
	}
	password := os.Getenv("PRODUCT_DB_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("PRODUCT_DB_PASSWORD is required")
	}
	name := os.Getenv("PRODUCT_DB_NAME")
	if name == "" {
		name = "products_db"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name), nil
}

func main() {
	dsn, err := databaseConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open product database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connect to product database: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS products (
			product_id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DOUBLE PRECISION NOT NULL CHECK (price > 0)
		)`)
	if err != nil {
		log.Fatalf("create products table: %v", err)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, &productServer{db: db})
	reflection.Register(grpcServer)

	log.Println("Product Service connected to PostgreSQL and running on port 50051")
	log.Println("Demo controls: PRODUCT_DELAY=2s PRODUCT_FAIL=true")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
