# Stage 1 - Build the application
FROM golang:1.26 AS builder
WORKDIR /app
COPY . .
RUN go build -o hello-service
# Stage 2 - Create a lightweight runtime image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/hello-service .
CMD ["./hello-service"]