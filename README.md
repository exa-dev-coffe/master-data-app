# Master Data Service (master-data)

Master Data Service is a core microservice built with **Go** and **Fiber**. It serves as the single source of truth for the system's core data, such as Menus, Categories, and Tables.

## 🚀 Technologies

- **Language**: Go 1.25
- **Framework**: Fiber v2
- **Database**: PostgreSQL
- **SQL Toolkit**: `sqlx`
- **Message Broker**: RabbitMQ (`amqp091-go`)
- **Storage**: MinIO (Object Storage SDK)
- **Observability**: OpenTelemetry (`otelsql`, OTLP gRPC)
- **Logging**: `log/slog` (JSON structured)

## 📦 Features

- **Menu Management**: CRUD operations for coffee and food menus.
- **Category Management**: Grouping items into specific categories.
- **File Uploads**: Direct integration with MinIO for product images.
- **Internal APIs**: Exposes secure HMAC-validated endpoints for inter-service communication (e.g., to `transaction-service`).
- **Distributed Tracing**: Automatic tracing of HTTP requests, SQL queries, and AMQP messages.

## 🛠️ Prerequisites

- Go 1.25+
- PostgreSQL
- MinIO
- RabbitMQ

## ⚙️ Environment Variables

Copy the `.env.example` file to `.env`:

```bash
cp .env.example .env
```

## 🚀 How to Run

1.  **Download Dependencies:**

    ```bash
    go mod download
    ```

2.  **Run Locally:**
    ```bash
    go run main.go
    ```

## 🐳 Docker Support

```bash
docker build -t eka-dev/master-data .
```

- **Storage**: MinIO (Object Storage SDK)
- **Observability**: OpenTelemetry (`otelsql`, OTLP gRPC)
- **Logging**: `log/slog` (JSON structured)

## 📦 Features

- **Menu Management**: CRUD operations for coffee and food menus.
- **Category Management**: Grouping items into specific categories.
- **File Uploads**: Direct integration with MinIO for product images.
- **Internal APIs**: Exposes secure HMAC-validated endpoints for inter-service communication (e.g., to `transaction-service`).
- **Distributed Tracing**: Automatic tracing of HTTP requests, SQL queries, and AMQP messages.

## 🛠️ Prerequisites

- Go 1.25+
- PostgreSQL
- MinIO
- RabbitMQ

## ⚙️ Environment Variables

Copy the `.env.example` file to `.env`:

```bash
cp .env.example .env
```

## 🚀 How to Run

1.  **Download Dependencies:**

    ```bash
    go mod download
    ```

2.  **Run Locally:**
    ```bash
    go run main.go
    ```

## 🧪 Integration Testing

Jalankan perintah berikut untuk mengeksekusi integration test suite dengan `testcontainers-go`:

```bash
go test -v .
```

_Persyaratan:_ Docker Desktop/Daemon harus aktif.

## 🐳 Docker Support

```bash
docker build -t eka-dev/master-data .
```
