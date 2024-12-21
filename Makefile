PROJECT_DIR = $(CURDIR)
PROJECT_BIN = $(PROJECT_DIR)/bin

DATA_PATH := $(PROJECT_DIR)/internal/app/service/fetcher/data

GOLANGCI_TAG = 1.61.0
SQLC_PATH ?= configs/sqlc.yaml
GOLANGCI_LINT_BIN = $(PROJECT_BIN)/golangci-lint

.PHONY: all
all: build-notifier build-scraper build-worker build-fetcher

# build binary
.PHONY: build-notifier
build-notifier:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/notifier ./cmd/notifier

.PHONY: build-scraper
build-scraper:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/scraper ./cmd/scraper

.PHONY: build-worker
build-worker:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/worker ./cmd/worker

.PHONY: build-fetcher
build-fetcher:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/fetcher ./cmd/fetcher

# sqlc
.PHONY: install-sqlc
install-sqlc:
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: sqlc
sqlc: install-sqlc
	@echo "generating sqlc..."
	@$(shell go env GOPATH)/bin/sqlc generate -f $(SQLC_PATH)

# linter
.PHONY: install-lint
install-lint:
	@if [ ! -f $(GOLANGCI_LINT_BIN) ]; then \
			$(info "Downloading golangci-lint v$(GOLANGCI_TAG)") \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_BIN) v$(GOLANGCI_TAG); \
		fi

.PHONY: lint
lint: install-lint
	$(GOLANGCI_LINT_BIN) run ./... --config=./configs/golangci.yml


.PHONY: mock
mock:
	@echo "generating mocks..."
	@go generate ./...

.PHONY: generate
generate: sqlc mock

.PHONY: test
test:
	@go test -v --timeout=1m --covermode=count --coverprofile=coverage_tmp.out ./...
	@cat coverage_tmp.out | grep -v "_mock.go" > coverage.out

.PHONY: covearge-html
coverage-html:
	@go tool cover --html=coverage.out

.PHONY: covearge-func
coverage-func:
	@go tool cover --func=coverage.out

# Docker build
.PHONY: docker-build-notifier
docker-build-notifier:
	docker build --build-arg BINARY=notifier -t notifier .

.PHONY: docker-build-scraper
docker-build-scraper:
	docker build --build-arg BINARY=scraper -t scraper .

.PHONY: docker-build-worker
docker-build-worker:
	docker build --build-arg BINARY=worker -t worker .

.PHONY: docker-build-fetcher
docker-build-fetcher:
	docker build --build-arg BINARY=fetcher --build-arg DATA_PATH=$(DATA_PATH) -t fetcher -f Dockerfile.fetcher .

.PHONY: docker-build-all
docker-build-all: docker-build-notifier docker-build-scraper docker-build-worker

# Docker Compose
.PHONY: docker-compose-up
docker-compose-up: docker-build-all
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose down

.PHONY: docker-compose-up-fetcher
docker-compose-up-fetcher: docker-build-fetcher
	docker-compose -f docker-compose.fetcher.yml up --abort-on-container-exit
.PHONY: docker-compose-down-fetcher
docker-compose-down-fetcher:
	docker-compose -f docker-compose.fetcher.yml down