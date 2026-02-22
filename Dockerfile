# ── Stage 1: Build the SvelteKit frontend ────────────────────────────────────
FROM node:22-alpine AS frontend

WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

# ── Stage 2: Build the Go binary ─────────────────────────────────────────────
FROM golang:1.26-alpine AS backend

WORKDIR /app

# Download dependencies first (cached layer).
COPY go.mod go.sum ./
RUN go mod download

# Copy source.
COPY . .

# Copy the compiled frontend into the expected embed path.
COPY --from=frontend /app/web/build ./web/build

# Build the binary. CGo is not needed (modernc.org/sqlite is pure Go).
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o bin/juno ./cmd/server

# ── Stage 3: Minimal runtime image ───────────────────────────────────────────
FROM alpine:3.21

# ca-certificates is needed for outbound HTTPS (S3, SMTP, etc.).
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend /app/bin/juno ./juno

# Railway injects PORT; default to 8080 for local docker runs.
ENV SERVER_PORT=8080
EXPOSE 8080

ENTRYPOINT ["./juno"]
