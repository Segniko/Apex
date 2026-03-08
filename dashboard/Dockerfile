# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Receiver binary
RUN go build -o receiver cmd/server/main.go

# Production Stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/receiver .

# Set environment defaults
ENV PORT=8080
ENV DATABASE_URL=postgres://postgres:postgres@postgres:5432/apex?sslmode=disable

EXPOSE 8080

CMD ["./receiver"]
