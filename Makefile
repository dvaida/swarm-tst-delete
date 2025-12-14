.PHONY: build test lint clean

build:
	go build -o bin/swarm-indexer ./cmd/swarm-indexer

test:
	go test ./...

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

clean:
	rm -rf bin/
