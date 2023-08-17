# Build stage
FROM golang:1.18-alpine3.17 AS builder

WORKDIR /github.com/b0shka/learning-backend

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/app/main.go

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /github.com/b0shka/learning-backend/.bin/app .
COPY --from=builder /github.com/b0shka/learning-backend/templates/ ./templates/
COPY --from=builder /github.com/b0shka/learning-backend/configs/ ./configs/
COPY --from=builder /github.com/b0shka/learning-backend/.env .

EXPOSE 8000
CMD ["./app"]