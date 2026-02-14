# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary
# CGO_ENABLED=0 disables CGo, making the binary statically linked
# -o specifies the output file name
# ./cmd/your-app is an example path to your main package
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o blackjack-tui .

# Stage 2: Create the final, lightweight image
FROM alpine:latest

# Install ca-certificates for HTTPS communication if needed
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add libc6-compat

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/blackjack-tui .
COPY . .

# Expose the port your application listens on (e.g., 8080)
EXPOSE 42069

# Command to run the application when the container starts
CMD ["./blackjack-tui", "server"]
