# Build stage
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o microcurrency .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/microcurrency .

# Copy the .env.example file
COPY --from=builder /app/.env.example ./.env.example

# Create data directory
RUN mkdir -p /data

# Expose the port
EXPOSE 8880

# Set environment variables
ENV PORT=8880
ENV DATA_DIR=/data

# Run the binary
CMD ["./microcurrency"]