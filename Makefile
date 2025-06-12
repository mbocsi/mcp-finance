build:
	@go build -o bin/mcp-finance

run: build
	@./bin/mcp-finance

test:
	@go test ./...
