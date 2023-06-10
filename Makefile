build:
	@echo "Building binary..."
	@go build -o bin/blocker .

run: build
	@echo "Running binary..."
	@./bin/blocker

test:
	@echo "Running tests..."
	@go test -v ./...