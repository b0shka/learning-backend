# Build stage
FROM golang:1.18-alpine3.17 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/app/main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/.bin/app .
COPY --from=builder /app/templates/ ./templates/
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/migrate ./migrate

# COPY .env .
COPY start.sh .
COPY wait-for.sh .
COPY internal/repository/postgresql/migration ./migration

EXPOSE 8080
CMD ["/app/app"]
ENTRYPOINT ["/app/start.sh"]