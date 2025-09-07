# -------- Build Stage --------
FROM golang:1.25-alpine AS builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory inside the container
WORKDIR /app

# Copy go mod files first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go app
RUN go build -o main .

# -------- Run Stage --------
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the built binary from builder
COPY --from=builder /app/main .

# Run the binary
CMD ["./main"]
