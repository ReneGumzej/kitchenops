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

# tzdata für korrekte Zeitzonen, ca-certificates für HTTPS-Aufrufe (Mistral-API),
# su-exec zum User-Wechsel im Entrypoint (ohne Passwort)
RUN apk add --no-cache tzdata ca-certificates su-exec && \
    # User und Gruppe für /data anlegen - UID/GID werden zur Laufzeit
    # über docker-compose.yml gesetzt (Standard: 1000:1000).
    # Falls nicht gesetzt, wird als Root laufen (nicht ideal für gemountete Volumes).
    addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /data

COPY --from=builder /out/api /usr/local/bin/api

# Entrypoint-Skript: stellt Schreibrechte auf /data sicher
# - Falls DB-Datei existiert: chmod 0666
# - Falls /data existiert: chmod 0777
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 3000

# Entrypoint läuft als root, um chmod/chown auf gemountete Volumes
# ausführen zu können. Das Skript wechselt dann zu appuser.
# In docker-compose.yml kann der User auf die Host-UID/GID gesetzt werden.
USER root

ENTRYPOINT ["/entrypoint.sh"]
CMD ["api"]
