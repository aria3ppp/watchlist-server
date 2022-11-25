# load .env file if exists
ENVFILE ?= .env
$(if $(wildcard $(ENVFILE)), \
	$(foreach VAR, $(shell sed -ne 's/ *\#.*$$//; /./ s/=.*$$// p' $(ENVFILE)), \
		$(if $($(VAR)),, \
			$(eval $(shell \
					echo export $(VAR)=$(shell sed -nr 's/$(VAR)=(.+)/\1/p' $(ENVFILE)) \
			)) \
		) \
	) \
)

MIGRATE_DSN ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"
MIGRATE := docker run --rm -v $(shell pwd)/migrations:/migrations --user "$(shell id -u):$(shell id -g)" --network host migrate/migrate -path=/migrations -database "$(MIGRATE_DSN)"

DOCKER_COMPOSE_SERVICES := docker compose -f docker-compose.services.yml
DOCKER_COMPOSE_SERVER := $(DOCKER_COMPOSE_SERVICES) -f docker-compose.server.yml

.PHONY: default
default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: services-ps
services-ps: ## list services containers
	$(DOCKER_COMPOSE_SERVICES) ps

.PHONY: services-up
services-up: ## create and start services
	$(DOCKER_COMPOSE_SERVICES) up -d

.PHONY: services-down
services-down: ## stop and remove services
	$(DOCKER_COMPOSE_SERVICES) down

.PHONY: server-ps
server-ps: ## list server containers
	$(DOCKER_COMPOSE_SERVER) ps

.PHONY: server-up
server-up: ## create and start server
	$(DOCKER_COMPOSE_SERVER) up -d

.PHONY: server-down
server-down: ## stop and remove server
	$(DOCKER_COMPOSE_SERVER) down

.PHONY: sync-sqlboiler-conf
sync-sqlboiler-conf: ## sync sqlboiler config file with passed envs
	@echo "Syncing sqlboiler config file with passed envs..."
	@sed -r -i 's/[[:space:]]*dbname[[:space:]]*=[[:space:]]*(.+)/dbname = "$(POSTGRES_DB)"/' sqlboiler.toml
	@sed -r -i 's/[[:space:]]*port[[:space:]]*=[[:space:]]*(.+)/port = $(POSTGRES_PORT)/' sqlboiler.toml
	@sed -r -i 's/[[:space:]]*user[[:space:]]*=[[:space:]]*(.+)/user = "$(POSTGRES_USER)"/' sqlboiler.toml
	@sed -r -i 's/[[:space:]]*pass[[:space:]]*=[[:space:]]*(.+)/pass = "$(POSTGRES_PASSWORD)"/' sqlboiler.toml

.PHONY: test-all
test-all: ## run all tests
	@echo "Running all tests..."
	@echo "mode: count" > coverage-all.out
	@make test-unit
	@tail -n +2 coverage.out >> coverage-all.out
	@make test-models
	@tail -n +2 coverage.out >> coverage-all.out
	@make test-repo-integration
	@tail -n +2 coverage.out >> coverage-all.out
	@make test-search-integration
	@tail -n +2 coverage.out >> coverage-all.out
	@make test-e2e
	@tail -n +2 coverage.out >> coverage-all.out
	@echo "All tests passed!"

.PHONY: test-unit
test-unit: ## run unit tests
	go test -covermode=count -coverprofile=coverage.out $(shell go list ./... | grep -v /models)

.PHONY: test-models
test-models: ## run models tests
	@make migrate
	go test -covermode=count -coverprofile=coverage.out  ./internal/models/...
	@make migrate-drop

.PHONY: test-repo-integration
test-repo-integration: ## run repository integration tests
	env TEST_DB_INTEGRATION=V go test -covermode=count -coverprofile=coverage.out ./internal/repo/

.PHONY: test-search-integration
test-search-integration: ## run search integration tests
	env TEST_ES_INTEGRATION=V go test -covermode=count -coverprofile=coverage.out ./internal/search/

.PHONY: test-e2e
test-e2e: ## run end-to-end tests
	env TEST_E2E=V go test -covermode=count -coverprofile=coverage.out ./internal/server/

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

.PHONY: test-e2e-cover
test-e2e-cover: test-e2e ## run end-to-end tests and show test coverage information
	go tool cover -html=coverage.out

.PHONY: test-arg-cover
test-arg-cover: ## run tests by passing $ARG env value to 'go test' command and show test coverge information
	go test -covermode=count -coverprofile=coverage.out $(ARG)
	go tool cover -html=coverage.out

.PHONY: run
run: ## build server and then run entrypoint.sh
	go build -o server .
	./entrypoint.sh

.PHONY: install-dev-binaries
install-dev-binaries: ## install command-line binaries used in 'go generate' and Makefile
	go install github.com/abice/go-enum@latest
	go install github.com/golang/mock/mockgen@v1
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

.PHONY: generate
generate: ## run 'go generate' for all packages
	go generate ./...

.PHONY: generate-models
generate-models: ## run sqlboiler to generate models
	sqlboiler --config sqlboiler.toml --output internal/models --no-auto-timestamps --wipe psql

.PHONY: lint
lint: ## run staticcheck
	@staticcheck ./...

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
	@echo "Resetting database..."
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
