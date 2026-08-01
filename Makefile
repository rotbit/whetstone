GOCTL := goctl
STYLE := go_zero

.PHONY: gen gen-user-api gen-user-rpc gen-interview-api gen-interview-rpc gen-question-rpc install-tools up down tidy

## 一键生成全部 go-zero 代码（首次克隆后执行）
gen: gen-user-api gen-user-rpc gen-interview-api gen-interview-rpc gen-question-rpc tidy

gen-user-api:
	cd app/user/api && $(GOCTL) api go -api user.api -dir . -style $(STYLE)

gen-user-rpc:
	cd app/user/rpc && $(GOCTL) rpc protoc user.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style $(STYLE)

gen-interview-api:
	cd app/interview/api && $(GOCTL) api go -api interview.api -dir . -style $(STYLE)

gen-interview-rpc:
	cd app/interview/rpc && $(GOCTL) rpc protoc interview.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style $(STYLE)

gen-question-rpc:
	cd app/question/rpc && $(GOCTL) rpc protoc question.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style $(STYLE)

tidy:
	go mod tidy

## 安装 goctl 及 protoc 插件
install-tools:
	go install github.com/zeromicro/go-zero/tools/goctl@latest
	$(GOCTL) env check --install --verbose --force

## 本地基础设施（MySQL/Redis/Qdrant/MinIO）
up:
	docker compose -f deploy/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose.yml down
