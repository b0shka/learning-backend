# learning-backend

![Go](https://img.shields.io/static/v1?label=GO&message=v1.18&color=blue)

---

## Installation

#### Prerequisites

- Go 1.18
- Docker & Docker Compose
- mockgen (used to start mock generation for unit tests)
- sqlc (used to run code generation for working with PostgreSQL)
- golang-migrate (used to run migrations in the database)
- golangci-lint (used to run code checks)
- swag (used to re-generate swagger documentation)

Create `.env` file in root directory and add following values:

```
MONDO_URI=mongodb://mongodb:27017
MONGO_DB_NAME=<db name>

POSTGRESQL_URL=postgresql://<user>:<password>@<host>:<port>/<db name>?sslmode=disable
MIGRATION_URL=file://<path to folder with migrate files>

REDIS_ADDRESS=<host>:<port>

EMAIL_SERVICE_NAME=<name complany>
EMAIL_SERVICE_ADDRESS=<email address>
EMAIL_SERVICE_PASSWORD=<email password>

SECRET_KEY=<random string>
CODE_SALT=<random string>

HTTP_HOST=localhost
```

---

## Build & Run

To start, run

```
make start
```