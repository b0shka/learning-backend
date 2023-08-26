PROGRAM_NAME = main
REGISTRY = b0shka
API_IMAGE = backend
POSTGRES_IMAGE = postgres
TAG = latest
NETWOTK=backend-network
DB_URL=postgresql://root:qwerty@localhost:5432/service?sslmode=disable
MIGRATION_URL=internal/repository/postgresql/migration

.PHONY: build start test lint swag mock network docker-build docker-run docker-push docker-run-postgres migrateup migratedown sqlc
.DEFAULT_GOAL := start

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/${PROGRAM_NAME} ./cmd/app/main.go

start: build
#	docker compose up
	APP_ENV="local" .bin/main

test:
#	make docker-run-postgres
#	make migrateup
	GIN_MODE=release go test --short -coverprofile=cover.out -v ./...
	make test.coverage
#	make migratedown
#	docker stop ${POSTGRES_IMAGE}
#	docker rm ${POSTGRES_IMAGE}

test.coverage:
	go tool cover -func=cover.out | grep "total"

lint:
	golangci-lint run

swag:
	swag init -g internal/app/app.go

mock:
	mockgen -source=internal/repository/mongodb/repository.go -destination=internal/repository/mongodb/mocks/mock_repository.go
	mockgen -destination=internal/repository/postgresql/mocks/mock_repository.go github.com/b0shka/backend/internal/repository/postgresql/sqlc Store
	mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_service.go

network:
	docker network create ${NETWOTK}

docker-build:
	docker build -f Dockerfile -t ${REGISTRY}/${API_IMAGE}:${TAG} .

docker-run:
	docker run --name ${API_IMAGE} --network ${NETWOTK} -p 8080:8080 -e GIN_MODE=release -e APP_ENV=local --rm -d ${REGISTRY}/${API_IMAGE}:${TAG}

docker-push:
	docker push ${REGISTRY}/${API_IMAGE}:${TAG}

docker-run-postgres:
	docker run --name ${POSTGRES_IMAGE} --network ${NETWOTK} -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=qwerty -e POSTGRES_DB=service -d postgres:15-alpine

migrateup:
	migrate -path ${MIGRATION_URL} -database ${DB_URL} -verbose up

migratedown:
	migrate -path ${MIGRATION_URL} -database ${DB_URL} -verbose down

sqlc:
	sqlc generate