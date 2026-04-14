.PHONY: dev test benchmark migrate-up migrate-down lint help

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
    		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## 开发环境启动服务
	go run ./server/cmd/api

test: ## 跑单侧
	go test ./server/... -v -count=1

benchmark: ## 跑 benchmark 做性能测试
	go test -bench=. -run=^$ ./server/...

migrate-up: ## migrate up。通过 make migrate-up DB_URL=xxx 使用
	migrate -path ./server/migrations -database "$(DB_URL)" up

migrate-down: ## 使 migration down 一个版本。通过 make migrate-down DB_URL=xxx 使用
	migrate -path ./server/migrations -database "$(DB_URL)" up

lint: ## 代码检查
	golangci-lint run ./...