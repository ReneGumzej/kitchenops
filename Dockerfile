# ============================================================================
# KitchenOps Backend
# Multi-Stage-Build: gorm.io/driver/sqlite basiert auf mattn/go-sqlite3 und
# benötigt daher CGO - im Builder wird gcc mitinstalliert.
# ============================================================================

# --- Stage 1: Bauen ---
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache build-base

# Plattform-Argumente mit sinnvollen Defaults (für Docker-Container: linux/amd64)
ARG GOOS=linux
ARG GOARCH=amd64

WORKDIR /src

# Abhängigkeiten zuerst cachen
COPY go.mod go.sum ./
RUN go mod download

# Quellcode kopieren und bauen
# Für Multi-Plattform-Builds: docker build --build-arg GOOS=darwin --build-arg GOARCH=arm64 ...
COPY . .
RUN CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o /out/api ./cmd/api

# --- Stage 2: Laufzeit-Image ---
FROM alpine:3.21

# tzdata für korrekte Zeitzonen, ca-certificates für HTTPS-Aufrufe (Mistral-API)
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /data

COPY --from=builder /out/api /usr/local/bin/api

# SQLite-Datenbank liegt im Volume /data (CWD der App)
EXPOSE 3000

ENTRYPOINT ["api"]
