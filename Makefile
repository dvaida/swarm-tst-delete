.PHONY: build test lint clean

build:
	go build -o bin/swarm-indexer ./cmd/swarm-indexer

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
