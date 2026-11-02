#!/bin/sh
# KitchenOps Entrypoint-Skript
# Stellt sicher, dass die SQLite-Datenbank im gemounteten Volume /data
# vom Container-User (appuser, UID 1000) beschrieben werden kann.
#
# Hintergrund: In Docker sind Host-UIDs und Container-UIDs isoliert.
# Selbst wenn beide UID 1000 haben, sind es unterschiedliche Users.
# Die Lösung: Berechtigungen explizit anpassen (als root) und dann
# auf den Ziel-User wechseln.

set -e

DATA_DIR="/data"
DB_FILE="${DATA_DIR}/kitchenops.db"

# 1. Verzeichnis-Rechte anpassen (als root)
if [ -d "${DATA_DIR}" ]; then
    echo "[entrypoint] Setze Berechtigungen für ${DATA_DIR}..."
    chmod 0777 "${DATA_DIR}" || echo "[entrypoint] Warnung: chmod auf ${DATA_DIR} fehlgeschlagen"
fi

# 2. Datenbank-Datei-Rechte anpassen (als root)
if [ -f "${DB_FILE}" ]; then
    echo "[entrypoint] Setze Berechtigungen für ${DB_FILE}..."
    chmod 0666 "${DB_FILE}" || echo "[entrypoint] Warnung: chmod auf ${DB_FILE} fehlgeschlagen"
fi

# 3. Wechsle zum appuser (UID 1000) für den Application-Start
# Falls in docker-compose.yml ein anderer User gesetzt wurde, wird
# dieser hier bereits aktiv sein und der Fall-through zum exec ist sicher.
if [ "$(id -u)" = "0" ]; then
    echo "[entrypoint] Wechsle von root zu appuser..."
    exec su-exec appuser "$@"
else
    echo "[entrypoint] Läuft bereits als nicht-root User, starte direkt..."
    exec "$@"
fi
