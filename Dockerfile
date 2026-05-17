# --- Stage 1: Build ---
FROM golang:1.26-alpine AS builder

# Install git and certificates (needed for downloading modules)
RUN apk add --no-cache git ca-certificates

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary (better for alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# --- Stage 2: Run ---
FROM alpine:latest

# Add ca-certificates for secure database connections (SSL/TLS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the .env file (Optional: usually handled by Docker Compose)
# COPY .env . 

# Expose the port your app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]