# Stage 1: Build Go binary
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

# Copy go mod dulu biar cache efisien
COPY go.mod go.sum ./
RUN go mod download

# Copy semua source code
COPY . .

# Build binary
RUN go build -o master-data .

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary
COPY --from=builder /app/master-data .

# Copy migrations biar bisa dijalankan
COPY db/migrations ./db/migrations

# Expose port aplikasi
EXPOSE 8080

CMD ["./master-data"]
