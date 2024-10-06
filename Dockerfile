FROM golang:1.22.8 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .

COPY --from=builder /app/.env .env
COPY --from=builder /app/db db
EXPOSE 8080
CMD ["./app"]

