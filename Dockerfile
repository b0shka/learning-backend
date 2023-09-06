# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/main ./cmd/app/main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/.bin/main ./main
COPY --from=builder /app/templates/ ./templates/
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/docs/ ./docs/
COPY internal/repository/postgresql/migration/ ./internal/repository/postgresql/migration
COPY .env .
COPY wait-for.sh .

EXPOSE 8080
CMD ["/app/main"]