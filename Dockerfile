# Use a multi-stage build to reduce image size
FROM golang:1.22-alpine AS builder

WORKDIR /myapp

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o ./bin/api ./cmd/api \
    && go build -ldflags '-w -s' -a -o ./bin/migrate ./cmd/migrate

# Create a smaller final image
FROM alpine:latest

WORKDIR /myapp

# Copy the binaries from the builder stage
COPY --from=builder /myapp/bin/api ./bin/api
COPY --from=builder /myapp/bin/migrate ./bin/migrate

# Expose the port and run the API
CMD ["./bin/api"]
EXPOSE 8080
