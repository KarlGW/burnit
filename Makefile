DB_PASSWORD := ""

.PHONY: test
test:
	@go clean -testcache
	@go vet ./... && go test ./... -race -short

.PHONY: test-full
test-full: start-database-postgres test-db stop-database-postgres


.PHONY: test-db
test-db:
	@bash -c "if [ -z $$DB_PASSWORD ]; then echo 'no password set'; exit 1; fi"
	@export DB_PASSWORD=$$DB_PASSWORD
	@go clean -testcache
	@go vet ./... && go test ./... -race

.PHONY: lint
lint:
	@golangci-lint run ./...


.PHONY: start-database-postgres
start-database-postgres:
	@bash -c "if [ -z $$DB_PASSWORD ]; then echo 'no password set'; exit 1; fi"	
	@docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=$$DB_PASSWORD -e POSTGRES_DB=burnit -d postgres:18 >/dev/null 2>&1

.PHONY: stop-database-postgres
stop-database-postgres:
	@docker stop postgres >/dev/null 2>&1 && docker rm postgres >/dev/null 2>&1
