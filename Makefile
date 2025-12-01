.PHONY: test
test:
	go vet ./... &&	go test ./... -race

.PHONY: lint
lint:
	golangci-lint run ./...
