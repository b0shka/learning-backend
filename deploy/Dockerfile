# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/main ./cmd/app/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/.bin/main ./main
COPY --from=builder /app/templates/ ./templates/
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/docs/ ./docs/

EXPOSE 8080
CMD ["/app/main"]