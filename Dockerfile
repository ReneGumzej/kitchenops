# ============================================================================
# KitchenOps Backend
# Multi-Stage-Build: gorm.io/driver/sqlite basiert auf mattn/go-sqlite3 und
# benötigt daher CGO - im Builder wird gcc mitinstalliert.
# ============================================================================

# --- Stage 1: Bauen ---
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /src

# Abhängigkeiten zuerst cachen
COPY go.mod go.sum ./
RUN go mod download

# Quellcode kopieren und bauen
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

# --- Stage 2: Laufzeit-Image ---
FROM alpine:3.21

# tzdata für korrekte Zeitzonen, ca-certificates für HTTPS-Aufrufe (Mistral-API)
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /data

COPY --from=builder /out/api /usr/local/bin/api

# SQLite-Datenbank liegt im Volume /data (CWD der App)
EXPOSE 3000

ENTRYPOINT ["api"]
