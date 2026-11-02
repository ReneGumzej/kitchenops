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

# Quellcode kopieren und bauen (native Architektur des Build-Hosts)
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /out/api ./cmd/api

# --- Stage 2: Laufzeit-Image ---
FROM alpine:3.21

# tzdata für korrekte Zeitzonen, ca-certificates für HTTPS-Aufrufe (Mistral-API)
RUN apk add --no-cache tzdata ca-certificates && \
    # User und Gruppe für /data anlegen - UID/GID werden zur Laufzeit
    # über docker-compose.yml gesetzt (Standard: 1000:1000).
    # Falls nicht gesetzt, wird als Root laufen (nicht ideal für gemountete Volumes).
    addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /data

COPY --from=builder /out/api /usr/local/bin/api

# SQLite-Datenbank liegt im Volume /data (CWD der App)
# Data-Verzeichnis gehört appuser (UID/GID 1000) für Schreibrechte auf gemountete Volumes
RUN chown appuser:appgroup /data

EXPOSE 3000

# Standard-User: appuser (UID 1000, GID 1000)
# In docker-compose.yml kann der User auf die Host-UID/GID gesetzt werden.
USER appuser

ENTRYPOINT ["api"]
