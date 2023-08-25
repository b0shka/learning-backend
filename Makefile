PROGRAM_NAME = app
REGISTRY = b0shka
API_IMAGE = backend
TAG = latest

.PHONY: build start test lint swag mock docker-build docker-run docker-push docker-run-postgres createdb dropdb migrateup migratedown sqlc
.DEFAULT_GOAL := start

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/${PROGRAM_NAME} ./cmd/app/main.go

start: build
	APP_ENV="local" .bin/app
# docker compose up

test:
	GIN_MODE=release go test --short -coverprofile=cover.out -v ./...
	make test.coverage

test.coverage:
	go tool cover -func=cover.out | grep "total"

lint:
	golangci-lint run

swag:
	${HOME}/go/bin/swag init -g internal/app/app.go

mock:
	mockgen -source=internal/repository/mongodb/repository.go -destination=internal/repository/mongodb/mocks/mock_repository.go
	mockgen -source=internal/repository/postgresql/sqlc/querier.go -destination=internal/repository/postgresql/mocks/mock_repository.go
	mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_service.go

docker-build:
	docker build -f Dockerfile -t ${REGISTRY}/${API_IMAGE}:${TAG} .

docker-run:
	docker run -d -p 8080:8080 -e GIN_MODE=release --rm --name ${API_IMAGE} ${REGISTRY}/${API_IMAGE}:${TAG}

docker-push:
	docker push ${REGISTRY}/${API_IMAGE}:${TAG}

docker-run-postgres:
	docker run --name postgres15 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=qwerty -d postgres:15-alpine

createdb:
	docker exec -it postgres15 createdb --username=root --owner=root service

dropdb:
	docker exec -it postgres15 dropdb service

migrateup:
	migrate -path internal/repository/postgresql/migration -database "postgresql://root:qwerty@localhost:5432/service?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/repository/postgresql/migration -database "postgresql://root:qwerty@localhost:5432/service?sslmode=disable" -verbose down

sqlc:
	sqlc generate