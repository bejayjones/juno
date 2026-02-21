.PHONY: run build web test test-v test-pkg lint clean dev

# Start the Go server (uses previously-built frontend or empty build/ dir).
run:
	go run ./cmd/server

# Build the frontend then the Go binary (production).
build: web
	go build -o bin/juno ./cmd/server

# Build the SvelteKit frontend into web/build/.
web:
	cd web && npm run build

# Install frontend dependencies.
web-install:
	cd web && npm install

# Start the SvelteKit dev server (proxies /api to Go on :8080).
dev:
	cd web && npm run dev

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
	rm -rf web/build
