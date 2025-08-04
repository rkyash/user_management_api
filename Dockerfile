FROM golang:1.23.1 AS builder

WORKDIR /app

# Install required dependencies
RUN apt-get update && apt-get install -y git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api .
COPY --from=builder /app/config/config.yaml ./config/
COPY --from=builder /app/statics/index.html ./statics/
COPY --from=builder /app/docs/* ./docs/

# Create logs directory
RUN mkdir -p /app/logs

EXPOSE 8080

CMD ["./api"]
