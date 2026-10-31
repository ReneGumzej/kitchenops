# KitchenOps

Küchenverwaltung mit Go-Backend (Fiber + SQLite), KI-Rezept-Generierung und
nativem HTML/CSS/JS-Frontend. Die LLM-Anbindung ist **provider-agnostisch**
über das [GoAI-Framework](https://goai.sh) - unterstützt Mistral, OpenAI,
Anthropic und Google.

## 🤖 LLM-Anbieter konfigurieren

Anbieter, Key und Modell werden über Umgebungsvariablen (bzw. die `.env`)
gesetzt - die App entscheidet selbst, welcher Anbieter aktiv ist:

```bash
# Variante 1: nur den Key des gewünschten Anbieters setzen,
# Anbieter wird automatisch erkannt
MISTRAL_API_KEY=mstrl_...            # oder
OPENAI_API_KEY=sk-...                 # oder
ANTHROPIC_API_KEY=sk-ant-...          # oder
GEMINI_API_KEY=...                    # (Google)

# Variante 2: Anbieter und Modell explizit
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o
```

| Anbieter | API-Key | Standardmodell |
|----------|---------|----------------|
| Mistral (Standard) | `MISTRAL_API_KEY` | `mistral-small-2603` |
| OpenAI | `OPENAI_API_KEY` | `gpt-4o` |
| Anthropic | `ANTHROPIC_API_KEY` | `claude-sonnet-4-5` |
| Google | `GEMINI_API_KEY` / `GOOGLE_GENERATIVE_AI_API_KEY` | `gemini-2.5-flash` |

Alle KI-Funktionen (Rezept-Generierung, Kategorisierung, Freitext-Erfassung)
nutzen strukturierte Ausgaben (JSON-Schema) und laufen identisch über alle
Anbieter.

## 🚀 Deployment mit Docker (Heimnetzwerk)

Der komplette Stack läuft mit einem Befehl und ist im Heimnetzwerk unter
**http://kitchenops.local** (und http://192.168.178.74) erreichbar:

```bash
docker compose up -d --build
```

- **frontend** (nginx): Port 80, liefert die statische App und proxied die
  API-Routen ans Backend (same-origin, kein CORS nötig)
- **backend** (Go): SQLite-Datenbank liegt im Volume `./data/` (persistent)
- **MISTRAL_API_KEY** wird automatisch aus der `.env` im Projektverzeichnis gelesen

Stoppen / Logs / Aktualisieren:

```bash
docker compose down          # Stack stoppen
docker compose logs -f      # Logs verfolgen
docker compose up -d --build # Nach Code-Änderungen neu bauen und starten
```

### Name "kitchenops.local" im Heimnetz auflösen

Der Docker-Stack veröffentlicht nur den Port - die Namensauflösung läuft über
mDNS (Avahi). Für dauerhafte Auflösung (überlebt Neustarts) einmalig als root
anlegen:

```bash
sudo tee /etc/systemd/system/kitchenops-mdns.service > /dev/null <<'EOF'
[Unit]
Description=KitchenOps mDNS-Alias kitchenops.local
After=avahi-daemon.service
Requires=avahi-daemon.service

[Service]
ExecStart=/usr/bin/avahi-publish -a -R kitchenops.local 192.168.178.74
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl enable --now kitchenops-mdns.service
```

(IP bei Bedarf anpassen.) Alternativen: DNS-Eintrag im Router
(kitchenops.local -> 192.168.178.74) oder auf einzelnen Geräten ein
`/etc/hosts`-Eintrag.

## 💻 Entwicklung ohne Docker

```bash
# Backend
go run ./cmd/api                              # läuft auf :3000 (.env wird geladen)

# Frontend (Terminal)
cd frontend && python3 -m http.server 5500    # http://localhost:5500
```

Dev-Server auf Port 5500x nutzen automatisch `http://localhost:3000` als
Backend-URL; die Docker-nginx-Variante nutzt den Reverse-Proxy (gleicher Host).
Einstellungen > Server-URL überschreibt beides manuell.

## 📁 Projektstruktur

```
kitchen-ops/
├── cmd/api/main.go       # Einstieg: CORS, Routen, .env-Load
├── pkg/api/              # HTTP-Handler (Lebensmittel, Einkaufsliste, Rezepte, KI)
├── pkg/llm/              # Mistral-Client (Rezept-Generierung + Batch-Kategorisierung)
├── pkg/models/           # GORM-Modelle + Regel-Klassifikator
├── pkg/schema/           # Request/Response-Schemas
├── pkg/database/         # SQLite-Engine
├── Dockerfile            # Backend (Multi-Stage, CGO für SQLite)
├── docker-compose.yml    # backend + frontend
├── data/                 # SQLite-Datenbank der Docker-Instanz (persistent)
└── frontend/             # HTML/CSS/JS-App + nginx.conf (Reverse-Proxy)
```

## 🔌 API-Endpunkte

| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/lebensmittel` | Alle Lebensmittel (mit Kategorien) |
| GET/POST/PUT/DELETE | `/lebensmittel[/:id]` | CRUD |
| POST | `/lebensmittel/kategorisieren` | Kategorien per Regeln + KI-Batch zuordnen |
| GET/POST/PUT/DELETE | `/einkaufsliste[/:id]` | Einkaufsliste (in SQLite persistiert) |
| GET/POST | `/rezepte` | Gespeicherte Rezepte / Rezept speichern |
| POST | `/rezepte/generieren` | KI-Rezepte auf Basis des Inventars (mistral-small-2603, structured output) |

Details zum Frontend siehe `frontend/README.md`.
