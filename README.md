# Ecommerce Microservices

A Go-based ecommerce system using gRPC between services and PostgreSQL for product and order data.

## Services

| Service | Address |
|---|---|
| Product | `localhost:50051` |
| Order | `localhost:50052` |
| Customer | `localhost:50053` |
| Inventory | `localhost:50054` |
| Payment | `localhost:50055` |
| Frontend | `http://localhost:8081` |

The protobuf definitions are in [`proto/`](proto/).

## Prerequisites

- Go 1.25 or newer
- Docker and Docker Compose
- WSL is supported and used by the project setup

## Start the databases

From the repository root, start PostgreSQL:

```bash
docker compose up -d product-db order-db
```

Check the database containers:

```bash
docker compose ps
```

## Run the services locally

Open a separate WSL terminal for each service. Run these commands from the repository root:

```bash
PRODUCT_DB_PASSWORD=change_this_product_password \\
PRODUCT_DB_HOST=localhost PRODUCT_DB_PORT=5433 \\
go run ./product-service
```

```bash
ORDER_DB_PASSWORD=change_this_order_password \\
ORDER_DB_HOST=localhost ORDER_DB_PORT=5434 \\
go run ./order-service
```

The other services can be started with:

```bash
go run ./customer-service
go run ./inventory-service
go run ./payment-service
```

Start the frontend in another terminal:

```bash
go run ./frontend
```

Open `http://localhost:8081` in a browser.

## Configuration

Copy `.env.example` as a reference for database settings. The product service requires `PRODUCT_DB_PASSWORD`; the order service requires its corresponding database password. When running services directly in WSL, use `localhost` with ports `5433` and `5434`. The Compose service names `product-db` and `order-db` are used only from inside the Docker network.

For product-service demo behavior, you can set:

```bash
PRODUCT_DELAY=2s
PRODUCT_FAIL=true
```

## Build

Build the Go module with:

```bash
go build ./...
```

The repository also contains a root [`Dockerfile`](Dockerfile). The current [`compose.yaml`](compose.yaml) provisions the PostgreSQL databases.

## Stop the databases

```bash
docker compose down
```

To also remove the PostgreSQL data volumes:

```bash
docker compose down -v
```

## API testing

The services expose gRPC APIs defined in `proto/ecommerce.proto` and `proto/inventory.proto`. Use Postman, `grpcurl`, or another gRPC client to test the endpoints. Export the Postman collection into the repository when available, for example under `postman/`.
