# Use a Golang base image
FROM golang:1.22 AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o sso ./cmd/sso/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o migrator ./cmd/migrator/main.go

# Use a lightweight base image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary files from the builder stage
COPY --from=builder /app/sso /app/migrator ./

# Set environment variables for database connection

# Expose any necessary ports
# EXPOSE <port_number>

# Run the migrator before starting the application
RUN ./migrator --storage-dsn="$STORAGE_DSN" --migrations-path=./migrations/postgres

# Command to run the application
CMD ["./sso"]
