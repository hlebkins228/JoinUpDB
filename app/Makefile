.PHONY: build migrate_up migrate_down migrate_create run_local run_cli stop_local swagger test integration_test install \
	run_db stop_db run_mongo stop_mongo mongo_seed run_server seed_db dicebear_images check_roles index_research

POSTGRES_CONTAINER := join_up_db
POSTGRES_IMAGE := postgres:16
POSTGRES_USER := postgres
POSTGRES_PASSWORD := postgres
POSTGRES_DB := join_up
POSTGRES_PORT := 5433
POSTGRES_HOST := localhost
POSTGRES_APP_USER := joinup_user
POSTGRES_APP_PASSWORD := user
POSTGRES_GUEST_PASSWORD := guest
POSTGRES_ADMIN_PASSWORD := admin

MONGO_CONTAINER := join_up_mongo
MONGO_IMAGE := mongo:7
MONGO_USER := mongo
MONGO_PASSWORD := mongo
MONGO_DB := join_up
MONGO_PORT := 27017
MONGO_HOST := localhost

POSTGRES_VOLUME := join_up_pgdata
MONGO_VOLUME := join_up_mongodata
MONGO_KEYFILE := /data/db/mongod.key

SERVER_PORT := 8080
SERVER_HOST := localhost
DOCKER_NET := join_up_net
PROD := false

MIGRATIONS_DIR := migrations
DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

ENVS := POSTGRES_USER=$(POSTGRES_USER) \
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	POSTGRES_HOST=$(POSTGRES_HOST) \
	POSTGRES_PORT=$(POSTGRES_PORT) \
	POSTGRES_DB=$(POSTGRES_DB) \
	MONGO_USER=$(MONGO_USER) \
	MONGO_PASSWORD=$(MONGO_PASSWORD) \
	MONGO_HOST=$(MONGO_HOST) \
	MONGO_PORT=$(MONGO_PORT) \
	MONGO_DB=$(MONGO_DB) \
	SERVER_PORT=$(SERVER_PORT) \
	SERVER_HOST=$(SERVER_HOST) \
	PROD=$(PROD)
	
run_server:
	@$(ENVS) POSTGRES_USER=$(POSTGRES_APP_USER) POSTGRES_PASSWORD=$(POSTGRES_APP_PASSWORD) go run ./cmd/server

run_cli:
	@go run ./cmd/cli

run_db:
	@docker network inspect $(DOCKER_NET) >/dev/null 2>&1 || docker network create $(DOCKER_NET)
	@docker start $(POSTGRES_CONTAINER) 2>/dev/null || docker run --name $(POSTGRES_CONTAINER) --network $(DOCKER_NET) \
		-v $(POSTGRES_VOLUME):/var/lib/postgresql/data \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 -d $(POSTGRES_IMAGE)

stop_db:
	@docker stop $(POSTGRES_CONTAINER)

run_mongo:
	@docker network inspect $(DOCKER_NET) >/dev/null 2>&1 || docker network create $(DOCKER_NET)
	@docker run --rm -v $(MONGO_VOLUME):/data/db $(MONGO_IMAGE) bash -c 'if [ ! -f $(MONGO_KEYFILE) ]; then openssl rand -base64 756 > $(MONGO_KEYFILE) && chmod 400 $(MONGO_KEYFILE); fi'
	@docker start $(MONGO_CONTAINER) 2>/dev/null || docker run --name $(MONGO_CONTAINER) --network $(DOCKER_NET) \
		-v $(MONGO_VOLUME):/data/db \
		-e MONGO_INITDB_ROOT_USERNAME=$(MONGO_USER) \
		-e MONGO_INITDB_ROOT_PASSWORD=$(MONGO_PASSWORD) \
		-e MONGO_INITDB_DATABASE=$(MONGO_DB) \
		-p $(MONGO_PORT):27017 -d $(MONGO_IMAGE) mongod --replSet rs0 --keyFile $(MONGO_KEYFILE)
	@sleep 3 && docker exec $(MONGO_CONTAINER) mongosh -u $(MONGO_USER) -p $(MONGO_PASSWORD) --authenticationDatabase admin --quiet --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27017"}]})' 2>/dev/null || true

stop_mongo:
	@docker stop $(MONGO_CONTAINER)

migrate_up:
	@$(ENVS) go run ./cmd/migrations up

migrate_down:
	@$(ENVS) go run ./cmd/migrations down

migrate_create:
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir $(MIGRATIONS_DIR) create $(NAME) sql

seed_db:
	@$(ENVS) POSTGRES_USER=joinup_admin POSTGRES_PASSWORD=$(POSTGRES_ADMIN_PASSWORD) python3 scripts/seed.py

dicebear_images:
	@$(ENVS) POSTGRES_USER=joinup_admin POSTGRES_PASSWORD=$(POSTGRES_ADMIN_PASSWORD) python3 scripts/dicebear_images.py

check_roles:
	@POSTGRES_CONTAINER=$(POSTGRES_CONTAINER) \
		POSTGRES_DB=$(POSTGRES_DB) \
		POSTGRES_GUEST_PASSWORD=$(POSTGRES_GUEST_PASSWORD) \
		POSTGRES_APP_PASSWORD=$(POSTGRES_APP_PASSWORD) \
		POSTGRES_ADMIN_PASSWORD=$(POSTGRES_ADMIN_PASSWORD) \
		./scripts/check_roles.sh

index_research:
	@$(ENVS) python3 scripts/index_research.py

install:
	go mod tidy
	go mod download


TEST_PACKAGES := ./internal/service

test:
	go test $(TEST_PACKAGES) -count=1 -v

integration_test:
	go test ./internal/integration_test -count=1 -v

swagger:
	@go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -d ./cmd/server,./internal/controller/handlers,./internal/controller/dto,./internal/models -o docs --parseGoList=false


.PHONY: allure-test allure-serve test-report test-report-serve

ALLURE_OUTPUT_PATH ?= $(CURDIR)
ALLURE_OUTPUT_FOLDER ?= allure-results
ALLURE_TEST_PACKAGES := ./internal/service ./internal/repository/postgres
ALLURE_VERSION := 2.46.1
ALLURE_CLI := .test-tools/node_modules/.bin/allure
TEST_REPORT_DIR := $(CURDIR)/test-report
TEST_REPORT_PORT ?= 9001

allure-test:
	python3 scripts/test_report.py prepare --results-dir "$(ALLURE_OUTPUT_PATH)/$(ALLURE_OUTPUT_FOLDER)" --report-dir "$(TEST_REPORT_DIR)"
	ALLURE_OUTPUT_PATH="$(ALLURE_OUTPUT_PATH)" ALLURE_OUTPUT_FOLDER="$(ALLURE_OUTPUT_FOLDER)" go test -tags allure $(ALLURE_TEST_PACKAGES) -run '^TestAllure' -count=1 -v -coverpkg=./internal/service,./internal/repository/postgres -coverprofile="$(TEST_REPORT_DIR)/coverage.out"
	python3 scripts/test_report.py build --results-dir "$(ALLURE_OUTPUT_PATH)/$(ALLURE_OUTPUT_FOLDER)" --report-dir "$(TEST_REPORT_DIR)"

$(ALLURE_CLI):
	npm install --prefix .test-tools --no-package-lock --no-audit --no-fund allure-commandline@$(ALLURE_VERSION)
	test -x $(ALLURE_CLI)
	touch $(ALLURE_CLI)

test-report: allure-test $(ALLURE_CLI)
	$(ALLURE_CLI) generate "$(ALLURE_OUTPUT_PATH)/$(ALLURE_OUTPUT_FOLDER)" --clean -o "$(TEST_REPORT_DIR)/allure"

test-report-server: test-report
	python3 -m http.server $(TEST_REPORT_PORT) --bind 127.0.0.1 --directory "$(TEST_REPORT_DIR)"
