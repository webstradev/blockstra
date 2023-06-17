build:
	@echo "Building binary..."
	@go build -o bin/blocker .

run: build
	@echo "Running binary..."
	@./bin/blocker

test:
	@echo "Running tests..."
	@go test -v ./...

proto:
	@echo "Generating protobuf definitions..."
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto
	@echo "Done generating!"

.PHONY: proto