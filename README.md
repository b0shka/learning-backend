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

POSTGRESQL_USER=<db user>
POSTGRESQL_PASSWORD=<db password>
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=5432
POSTGRESQL_DB_NAME=<db name>

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