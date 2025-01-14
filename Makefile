APP_NAME?=test-app
MIGRATION_FILE=migrations/1_init.up.sql

migrate:
	@echo "Running migration..."
	go run cmd/migrator/main.go -migration=$(MIGRATION_FILE)

run:
	go run cmd/api/main.go

clean:
	rm -f ${APP_NAME}

build: clean
	go build -o ${APP_NAME} ./cmd/api/main.go

test:
	go test -v -count=1 ./...

test100:
	go test -v -count=100 ./...

race:
	go test -v -race -count=1 ./...

.PHONY: cover
cover:
	go test -short -count=1 -coverprofile="coverage.out" ./...
	go tool cover -html="coverage.out"
	rm "coverage.out"

.PHONY: gen
gen:
	mockgen -source="internal/api/file/storage.go" -destination="internal/api/file/mocks/mock_file_repository.go"
	protoc --go_out=. --go_opt=paths=source_relative api/proto/fileservice.proto
	protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/fileservice.proto

docker-up:
	docker-compose up --build

