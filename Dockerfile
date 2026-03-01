# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api/main.go

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Install CA certificates for HTTPS requests (needed for Gemini/Supabase)
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/bin/api ./api

# Copy the openapi.yaml for Scalar documentation
COPY --from=builder /app/api/openapi.yaml ./api/openapi.yaml

# Expose the API port
EXPOSE 8080

# Run the application
CMD ["./api"]
