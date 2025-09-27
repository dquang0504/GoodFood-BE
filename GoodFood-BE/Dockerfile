# Stage 1: Build
FROM golang:1.23.4 AS builder

WORKDIR /app

# Install dependencies to utilize docker layer caching later
COPY go.mod go.sum ./
RUN go mod download

# Copy soure code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/api ./cmd/api/main.go

# Build Worker binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/worker ./cmd/worker/main.go

# Stage 2: Runtime
FROM alpine:3.18

WORKDIR /app

# Install mandatory libs (such as: db connection)
RUN apk add --no-cache ca-certificates

# Copy binary from builder stage
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/worker /app/worker

# Default run cmd is API
CMD ["/app/api"]