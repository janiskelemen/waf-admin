.PHONY: run build lint
build:
	go build ./...

run:
	go run ./cmd/waf-admin -config configs/config.example.yaml

lint:
	@echo "Add golangci-lint here"
