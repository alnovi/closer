.SILENT:

## lint: статический анализ
.PHONY: lint
lint:
	@go tool golangci-lint run ./...

## lint-fix: статический анализ (авто-исправление)
.PHONY: lint-fix
lint-fix:
	go tool golangci-lint run ./... --fix --timeout 650s

## test: запуск тестов
.PHONY: test
test:
	@go tool gotestsum --format=testname -- -count=1 -coverpkg=./... -coverprofile=./cover.out ./...
	@go tool cover -html=./cover.out
	@rm ./coverage.out