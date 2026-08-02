GO_BIN ?= $(shell go env GOPATH)/bin
GOCTL ?= $(shell command -v goctl 2>/dev/null || printf '%s/goctl' '$(GO_BIN)')
GOCTL_VERSION ?= v1.10.2
GO_ZERO_STYLE ?= go_zero

BIN_DIR ?= bin
SERVICE ?= app-apis
IMAGE ?= whetstone/$(SERVICE):local
GO_BUILD_FLAGS ?= -trimpath
DOCKER_SERVICES := app-apis user-rpc interview-rpc question-rpc report-worker

export PATH := $(GO_BIN):$(PATH)

APP_APIS_DIR := app/apis/cmd/app-apis
APP_APIS_DESC_DIR := $(APP_APIS_DIR)/desc
USER_RPC_DIR := app/user/rpc
INTERVIEW_RPC_DIR := app/interview/rpc
QUESTION_RPC_DIR := app/question/rpc
REPORT_WORKER_DIR := app/pump/cmd/report-worker

.DEFAULT_GOAL := help

.PHONY: help generate generate-all generate_all gen check-goctl install-tools \
	gen-app-apis gen-rpcs gen-user-rpc gen-interview-rpc gen-question-rpc \
	build build-all build-app-apis build-user-rpc build-interview-rpc \
	build-question-rpc build-report-worker docker-build docker-build-all \
	docker-build-app-apis docker-build-user-rpc docker-build-interview-rpc \
	docker-build-question-rpc docker-build-report-worker tidy test up down

help:
	@echo "Available targets:"
	@echo "  make generate                                  Generate all stubs"
	@echo "  make generate type=api name=app-apis          Generate app-apis"
	@echo "  make generate type=rpc name=user              Generate one RPC service"
	@echo "  make generate-all                             Generate app-apis and all RPCs"
	@echo "  make build type=<api|rpc|worker> name=<name>  Build one service"
	@echo "  make build-all                                Build all five binaries"
	@echo "  make docker-build SERVICE=user-rpc            Build one service image"
	@echo "  make docker-build-all                         Build all five images"
	@echo "  make test                                     Run all Go tests"

## Reference-project-compatible entry point for generating one service.
generate:
ifeq ($(type),api)
	@test "$(name)" = "app-apis" || { echo "api name must be app-apis"; exit 1; }
	$(MAKE) gen-app-apis
else ifeq ($(type),rpc)
	@test -n "$(name)" || { echo "rpc name is required"; exit 1; }
	$(MAKE) gen-$(name)-rpc
else ifeq ($(type),)
	$(MAKE) generate-all
else
	@echo "type must be api or rpc"
	@exit 1
endif

generate-all: check-goctl
	$(MAKE) gen-app-apis
	$(MAKE) gen-rpcs
	$(MAKE) tidy

generate_all: generate-all

gen: generate-all

gen-app-apis: check-goctl
	cd $(APP_APIS_DESC_DIR) && $(GOCTL) api go --api app.api --dir .. --style $(GO_ZERO_STYLE)
	@# app-api is the goctl-compatible spec name; runtime config stays app-apis.
	@perl -pi -e 's#etc/app-api\.yaml#etc/app-apis.yaml#g' $(APP_APIS_DIR)/app.go
	@rm -f $(APP_APIS_DIR)/etc/app-api.yaml

gen-rpcs: gen-user-rpc gen-interview-rpc gen-question-rpc

gen-user-rpc: check-goctl
	cd $(USER_RPC_DIR)/pb && $(GOCTL) rpc protoc user.proto \
		--go_out=.. --go-grpc_out=.. --zrpc_out=.. --style $(GO_ZERO_STYLE) -m

gen-interview-rpc: check-goctl
	cd $(INTERVIEW_RPC_DIR)/pb && $(GOCTL) rpc protoc interview.proto \
		--go_out=.. --go-grpc_out=.. --zrpc_out=.. --style $(GO_ZERO_STYLE) -m

gen-question-rpc: check-goctl
	cd $(QUESTION_RPC_DIR)/pb && $(GOCTL) rpc protoc question.proto \
		--go_out=.. --go-grpc_out=.. --zrpc_out=.. --style $(GO_ZERO_STYLE) -m

check-goctl:
	@test -x "$(GOCTL)" || { \
		echo "goctl not found at $(GOCTL). Run 'make install-tools' first."; \
		exit 1; \
	}

install-tools:
	GOBIN=$(GO_BIN) go install github.com/zeromicro/go-zero/tools/goctl@$(GOCTL_VERSION)
	$(GOCTL) env check --install --verbose --force

## Reference-project-compatible entry point for building one service.
build:
ifeq ($(type),api)
	@test "$(name)" = "app-apis" || { echo "api name must be app-apis"; exit 1; }
	$(MAKE) build-app-apis
else ifeq ($(type),rpc)
	@test -n "$(name)" || { echo "rpc name is required"; exit 1; }
	$(MAKE) build-$(name)-rpc
else ifeq ($(type),worker)
	@test "$(name)" = "report-worker" || { echo "worker name must be report-worker"; exit 1; }
	$(MAKE) build-report-worker
else
	@echo "type must be api, rpc, or worker"
	@exit 1
endif

build-all: build-app-apis build-user-rpc build-interview-rpc build-question-rpc build-report-worker

build-app-apis:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/app-apis ./$(APP_APIS_DIR)

build-user-rpc:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/user-rpc ./$(USER_RPC_DIR)

build-interview-rpc:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/interview-rpc ./$(INTERVIEW_RPC_DIR)

build-question-rpc:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/question-rpc ./$(QUESTION_RPC_DIR)

build-report-worker:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/report-worker ./$(REPORT_WORKER_DIR)

docker-build:
	@case " $(DOCKER_SERVICES) " in *" $(SERVICE) "*) ;; *) echo "unsupported SERVICE: $(SERVICE)"; exit 1 ;; esac
	docker build --target $(SERVICE) -t $(IMAGE) .

docker-build-all: docker-build-app-apis docker-build-user-rpc docker-build-interview-rpc docker-build-question-rpc docker-build-report-worker

docker-build-app-apis:
	$(MAKE) docker-build SERVICE=app-apis IMAGE=whetstone/app-apis:local

docker-build-user-rpc:
	$(MAKE) docker-build SERVICE=user-rpc IMAGE=whetstone/user-rpc:local

docker-build-interview-rpc:
	$(MAKE) docker-build SERVICE=interview-rpc IMAGE=whetstone/interview-rpc:local

docker-build-question-rpc:
	$(MAKE) docker-build SERVICE=question-rpc IMAGE=whetstone/question-rpc:local

docker-build-report-worker:
	$(MAKE) docker-build SERVICE=report-worker IMAGE=whetstone/report-worker:local

tidy:
	go mod tidy

test:
	go test ./...

up:
	docker compose -f deploy/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose.yml down
