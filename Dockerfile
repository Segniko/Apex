# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build a truly static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o receiver cmd/server/main.go

# Production Stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary and templates
COPY --from=builder /app/receiver .
COPY --from=builder /app/templates ./templates

# Set environment defaults
ENV PORT=8081
ENV DATABASE_URL=postgresql://root@cockroach:26257/defaultdb?sslmode=disable
ENV REDIS_URL=redis:6379

EXPOSE 8081

CMD ["./receiver"]
