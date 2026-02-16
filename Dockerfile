# Stage 1: The Builder (Compiles the code)
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 1. Copy the dependency files FIRST
# We do this separate from the source code so Docker can cache the downloads.
COPY go.mod go.sum ./

# 2. Download the libraries (go-redis)
RUN go mod download

# 3. Copy the rest of the source code
COPY . .

# 4. Build the binary (named "server")
# CGO_ENABLED=0 ensures a static binary that runs anywhere
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Stage 2: The Runner (Tiny image)
FROM alpine:latest
WORKDIR /root/

# Copy only the compiled binary from the Builder stage
COPY --from=builder /app/server .

# Run it
CMD ["./server"]