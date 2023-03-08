# load env file
ENVFILE ?= .env
include $(ENVFILE)
export

MIGRATE_DSN ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"
# As test-models target depend on migrate command, use the docker one-liner run command instead
# MIGRATE := migrate -path=migrations -database "$(MIGRATE_DSN)"
MIGRATE := docker run --rm -v $(shell pwd)/migrations:/migrations --user "$(shell id -u):$(shell id -g)" --network host migrate/migrate -path=/migrations -database "$(MIGRATE_DSN)"

DOCKER_COMPOSE_SERVICES := docker compose -f docker-compose.services.yml
DOCKER_COMPOSE_SERVER := $(DOCKER_COMPOSE_SERVICES) -f docker-compose.server.yml

APPEND_COVERAGE_CMD := tail -n +2 coverage.out >> coverage-all.out

.PHONY: default
default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: services-ps
services-ps: ## list services containers
	$(DOCKER_COMPOSE_SERVICES) ps

.PHONY: services-up
services-up: ## create and start services
	$(DOCKER_COMPOSE_SERVICES) up -d

.PHONY: services-down
services-down: ## stop and remove services
	$(DOCKER_COMPOSE_SERVICES) down

.PHONY: services-postgres-up
services-postgres-up: ## create and start postgres service
	$(DOCKER_COMPOSE_SERVICES) up --wait postgres

.PHONY: services-elasticsearch-up
services-elasticsearch-up: ## create and start elasticsearch service
	$(DOCKER_COMPOSE_SERVICES) up --wait elasticsearch

.PHONY: services-minio-up
services-minio-up: ## create and start minio service
	$(DOCKER_COMPOSE_SERVICES) up --wait minio

.PHONY: server-ps
server-ps: ## list server containers
	$(DOCKER_COMPOSE_SERVER) ps

.PHONY: server-up
server-up: ## create and start server
	$(DOCKER_COMPOSE_SERVER) up -d

.PHONY: server-down
server-down: ## stop and remove server
	$(DOCKER_COMPOSE_SERVER) down

.PHONY: test-all
test-all: ## run all tests
	@echo "Running all tests..."
	@echo "mode: count" > coverage-all.out
	@make test-unit
	@$(APPEND_COVERAGE_CMD)
	@make test-models
	# exclude models test coverage as it is generated
	# @$(APPEND_COVERAGE_CMD)
	@make test-repo-integration
	@$(APPEND_COVERAGE_CMD)
	@make test-search-integration
	@$(APPEND_COVERAGE_CMD)
	@make test-storage-integration
	@$(APPEND_COVERAGE_CMD)
	@make test-e2e
	@$(APPEND_COVERAGE_CMD)
	@echo "All tests passed!"

.PHONY: test-unit
test-unit: ## run unit tests
	go test -covermode=count -coverprofile=coverage.out $(shell go list ./... | grep -v /models)

.PHONY: test-models
test-models: sync-sqlboiler-conf ## run models tests
	@make migrate
	go test -covermode=count -coverprofile=coverage.out  ./internal/models/...
	@make migrate-drop

.PHONY: sync-sqlboiler-conf
sync-sqlboiler-conf: ## sync sqlboiler config file with passed envs
	@echo "Syncing sqlboiler config file with passed envs..."
	@sed -r -i 's/[[:space:]]*dbname[[:space:]]*=[[:space:]]*(.+)/dbname = "$(POSTGRES_DB)"/' sqlboiler.toml
	@sed -r -i 's/[[:space:]]*port[[:space:]]*=[[:space:]]*(.+)/port = $(POSTGRES_PORT)/' sqlboiler.toml
	@sed -r -i 's/[[:space:]]*user[[:space:]]*=[[:space:]]*(.+)/user = "$(POSTGRES_USER)"/' sqlboiler.toml
	@sed -r -i 's/[[:space:]]*pass[[:space:]]*=[[:space:]]*(.+)/pass = "$(POSTGRES_PASSWORD)"/' sqlboiler.toml

.PHONY: test-repo-integration
test-repo-integration: ## run repository integration tests
	env TEST_DB_INTEGRATION=V go test -covermode=count -coverprofile=coverage.out ./internal/repo/

.PHONY: test-search-integration
test-search-integration: ## run search integration tests
	env TEST_ES_INTEGRATION=V go test -covermode=count -coverprofile=coverage.out ./internal/search/

.PHONY: test-storage-integration
test-storage-integration: ## run storage integration tests
	env TEST_MINIO_INTEGRATION=V go test -covermode=count -coverprofile=coverage.out ./internal/storage/

.PHONY: test-e2e
test-e2e: ## run end-to-end tests
	env TEST_E2E=V go test -covermode=count -coverprofile=coverage.out ./internal/server/

.PHONY: test-arg
test-arg: ## run tests by passing $ARG env value to 'go test' command
	go test -covermode=count -coverprofile=coverage.out $(ARG)

.PHONY: test-all-cover
test-all-cover: test-all ## run all tests and show test coverage information
	go tool cover -html=coverage-all.out

.PHONY: test-unit-cover
test-unit-cover: test-unit ## run unit tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-models-cover
test-models-cover: test-models ## run models tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-repo-integration-cover
test-repo-integration-cover: test-repo-integration ## run repository integration tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-search-integration-cover
test-search-integration-cover: test-search-integration ## run search integration tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-storage-integration-cover
test-storage-integration-cover: test-storage-integration ## run storage integration tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-e2e-cover
test-e2e-cover: test-e2e ## run end-to-end tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-arg-cover
test-arg-cover: test-arg ## run tests by passing $ARG env value to 'go test' command and show test coverge information
	go tool cover -html=coverage.out

.PHONY: run
run: ## build server and then run entrypoint.sh
	go build -o server .
	./entrypoint.sh

.PHONY: install-dev-deps
install-dev-deps: ## install dev dependencies
	go install github.com/golang/mock/mockgen@v1
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	# go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

.PHONY: generate
generate: ## run 'go generate' for all packages
	go generate ./...

.PHONY: generate-models
generate-models: sync-sqlboiler-conf ## run sqlboiler to generate models
	sqlboiler --config sqlboiler.toml --output internal/models --no-auto-timestamps --wipe psql

.PHONY: lint
lint: ## run linters
	golangci-lint run ./...

.PHONY: migrate
migrate: ## run all new database migrations
	@echo "Running all new database migrations..."
	@$(MIGRATE) up

.PHONY: migrate-down
migrate-down: ## revert database to the last migration step
	@echo "Reverting database to the last migrate step..."
	@$(MIGRATE) down 1

.PHONY: migrate-drop
migrate-drop: ## drop all database migrations
	@echo "dropping database..."
	@$(MIGRATE) drop -f

.PHONY: migrate-new
migrate-new: ## create a new database migration
	@read -p "Enter the name of the new migration: " name; \
	$(MIGRATE) create -ext sql -dir /migrations/ $${name}

.PHONY: migrate-reset
migrate-reset: ## reset database and re-run all migrations
	@echo "Resetting database..."
	@$(MIGRATE) drop -f
	@echo "Running all database migrations..."
	@$(MIGRATE) up

.PHONY: migrate-arg
migrate-arg: ## run migration command with argument ARG
	@echo "Running migration command with argument: $(ARG)"
	@$(MIGRATE) $(ARG)
