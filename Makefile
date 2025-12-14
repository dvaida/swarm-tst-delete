.PHONY: build test lint clean

build:
	go build -o bin/swarm-indexer ./cmd/swarm-indexer

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
