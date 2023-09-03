PROGRAM_NAME = main
REGISTRY = service
API_IMAGE = backend
POSTGRES_IMAGE = postgres
REDIS_IMAGE=redis
NGINX_IMAGE=nginx
TAG = latest
DB_URL=postgresql://root:qwerty@localhost:5432/service?sslmode=disable
MIGRATION_URL=internal/repository/postgresql/migration

.PHONY: build start test lint swag mock docker-build docker-run docker-push docker-run-postgres docker-run-redis migrateup migratedown sqlc
.DEFAULT_GOAL := start

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o ./.bin/${PROGRAM_NAME} ./cmd/app/main.go

start:
	docker compose down
	docker rmi cr.selcloud.ru/${REGISTRY}/${API_IMAGE}:${TAG}
	docker rmi cr.selcloud.ru/${REGISTRY}/${NGINX_IMAGE}:${TAG}
	docker compose up --build --force-recreate
# APP_ENV="local" .bin/main

test:
#	make docker-run-postgres
#	make migrateup
	GIN_MODE=release go test --short -coverprofile=cover.out -v -count=1 ./...
	make test.coverage
#	docker stop ${POSTGRES_IMAGE}
#	docker rm ${POSTGRES_IMAGE}

test.coverage:
	go tool cover -func=cover.out | grep "total"

lint:
	golangci-lint run ./... --config=./.golangci.yml

swag:
	swag init -g internal/app/app.go

mock:
	mockgen -destination=internal/repository/postgresql/mocks/mock_repository.go github.com/b0shka/backend/internal/repository/postgresql/sqlc Store
	mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_service.go
	mockgen -destination internal/worker/mocks/mock_worker.go github.com/b0shka/backend/internal/worker TaskDistributor

docker-build:
	docker build -f Dockerfile -t cr.selcloud.ru/${REGISTRY}/${API_IMAGE}:${TAG} .

docker-run:
	docker run --name ${API_IMAGE} -p 8080:8080 -e GIN_MODE=release -e APP_ENV=local --rm -d cr.selcloud.ru/${REGISTRY}/${API_IMAGE}:${TAG}

docker-push:
	docker push cr.selcloud.ru/${REGISTRY}/${API_IMAGE}:${TAG}

docker-run-postgres:
	docker run --name ${POSTGRES_IMAGE} -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=qwerty -e POSTGRES_DB=service -d postgres:15-alpine

docker-run-redis:
	docker run --name ${REDIS_IMAGE} -p 6379:6379 -d redis:7-alpine

create-migration:
	migrate create -ext sql -dir ${MIGRATION_URL} -seq ${name}

migrateup:
	migrate -path ${MIGRATION_URL} -database ${DB_URL} -verbose up

migratedown:
	migrate -path ${MIGRATION_URL} -database ${DB_URL} -verbose down

sqlc:
	sqlc generate