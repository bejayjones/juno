.PHONY: run build test lint clean

run:
	go run ./cmd/server

build:
	go build -o bin/juno ./cmd/server

test:
	go test ./...

test-v:
	go test -v ./...

# Run tests for a single package, e.g.: make test-pkg PKG=./internal/inspection/...
test-pkg:
	go test $(PKG)

lint:
	golangci-lint run

clean:
	rm -rf bin/ data/ *.db *.db-shm *.db-wal
