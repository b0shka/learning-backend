# Build stage
FROM golang:1.18-alpine3.17 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/app/main.go

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/.bin/app .
COPY --from=builder /app/templates/ ./templates/
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./app"]