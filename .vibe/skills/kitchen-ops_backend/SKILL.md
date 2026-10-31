---
name: kitchen-ops
description: |
  Skill für die Arbeit mit dem KitchenOps-Projekt – einem Go-basierten Backend für Küchenverwaltung (Lebensmittel, Rezepte, Inventar) mit REST-API (Fiber) und PostgreSQL (GORM).
  Enthält Anweisungen für Code-Stil, Datenbank-Operationen, API-Entwicklung und MCP-Server-Integration.

# Metadaten
version: 1.0.0
author: Mistral Vibe
language: de
framework: Go (Fiber, GORM)
database: PostgreSQL
---

# KitchenOps – Go Backend für Küchenverwaltung

Dieser Skill definiert Richtlinien und Werkzeuge für die Entwicklung im **KitchenOps**-Projekt.

---

## 📌 Projektübersicht

**KitchenOps** ist ein Backend-System zur Verwaltung von:
- **Lebensmitteln** (Name, Anzahl, Einheit)
- **Rezepten** (mit Kategorien und Many-to-Many-Zuordnung zu Lebensmitteln)
- **Sonstigen Gegenständen** (ohne Nährwertangaben)

### Technologie-Stack
| Komponente       | Technologie          |
|------------------|----------------------|
| **Web-Framework**| [Fiber](https://gofiber.io/) (v3) |
| **Datenbank**    | PostgreSQL + [GORM](https://gorm.io/) (v1.31) |
| **Konfiguration**| [Koanf](https://github.com/knadh/koanf) (v2) |
| **MCP-Server**   | [mcp-go](https://github.com/mark3labs/mcp-go) (v0.58) |
| **Go-Version**   | 1.26.5 |

---

## 🎯 Anweisungen für Vibe

### 1. Code-Stil & Konventionen

#### Allgemein
- **Sprache**: Code-Kommentare und Dokumentation **auf Deutsch**.
- **Namen**: 
  - Variablen/Funktionen: `camelCase` (z. B. `getLebensmittelByID`)
  - Typen/Strukturen: `PascalCase` (z. B. `LebensmittelKategorie`)
  - Paketnamen: `kleingeschrieben` (z. B. `pkg/api`, `pkg/models`)
- **Fehlerbehandlung**: 
  - Immer `err != nil` prüfen und mit `fiber.NewError` oder `log.Fatal` behandeln.
  - **Nicht** ignorieren (aktuelle Handler haben diesen Fehler).

#### Datenbank (GORM)
- **Modelle**: Immer mit GORM-Tags für Typen und Constraints:
  ```go
  type Lebensmittel struct {
      ID        uint      `gorm:"primaryKey;autoIncrement"`
      Name      string    `gorm:"type:varchar(255);not null;index"`
      CreatedAt time.Time `gorm:"autoCreateTime"`
  }
  ```
- **Migrationen**: 
  - Immer `db.AutoMigrate(&Model1{}, &Model2{})` in `cmd/api/main.go` aufrufen.
  - **Nicht** manuelle SQL-Migrationen verwenden.

#### API (Fiber)
- **Handler**: 
  - Immer in `pkg/api/handler.go` als Methoden des `DatabaseHandler` implementieren.
  - **Request/Response-Schemas** in `pkg/schema/schema.go` definieren.
  - **Fehlercodes** korrekt verwenden:
    - `400 Bad Request` für Validierungsfehler.
    - `404 Not Found` für nicht gefundene Ressourcen.
    - `500 Internal Server Error` für Datenbankfehler.
- **Routing**: 
  - Endpunkte in `cmd/api/main.go` registrieren.
  - **REST-konform** benennen (z. B. `/lebensmittel`, `/rezepte`).

#### MCP-Server
- **Zweck**: KI-gestützte Abfragen (z. B. Rezeptvorschläge).
- **Implementierung**: 
  - Tools in `pkg/mcp/server.go` definieren (z. B. `GetRezeptVorschlaege`).
  - **MCP-Spec**: [Model Context Protocol](https://github.com/modelcontextprotocol/spec) beachten.

---

### 2. Entwicklungsworkflow

#### Datenbank-Änderungen
1. Modell in `pkg/models/records.go` anpassen.
2. Migration in `cmd/api/main.go` aktualisieren (`db.AutoMigrate`).
3. **Nicht vergessen**: 
   - Indizes für häufig abgefragte Felder hinzufügen (z. B. `index` für `Name`).
   - `CreatedAt`/`UpdatedAt` für alle Tabellen verwenden.

#### Neue API-Endpunkte
1. **Schema** in `pkg/schema/schema.go` definieren (z. B. `NewRezeptRecord`).
2. **Handler-Methode** in `pkg/api/handler.go` implementieren.
3. **Route** in `cmd/api/main.go` registrieren.
4. **Fehlerbehandlung** einbauen (siehe [Fehlerbehandlung](#fehlerbehandlung)).

#### Fehlerbehandlung
**❌ Falsch** (aktueller Code in `handler.go`):
```go
if err := ctx.Bind().Body(newLebensmittel); err != nil {
    fiber.NewErrorf(fiber.StatusNoContent, err) // ❌ Fehler wird nicht zurückgegeben!
}
```

**✅ Richtig**:
```go
if err := ctx.Bind().Body(newLebensmittel); err != nil {
    return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
}

result := handler.db.Create(&lebensmittelTable)
if result.Error != nil {
    return fiber.NewError(fiber.StatusInternalServerError, "Datenbankfehler: "+result.Error.Error())
}
```

---

### 3. Wichtige Dateien & Verantwortlichkeiten

| Datei | Verantwortung |
|-------|---------------|
| `cmd/api/main.go` | **API-Einstiegspunkt**: Server-Initialisierung, Routing, DB-Verbindung. |
| `pkg/database/db.go` | **Datenbank-Konfiguration**: PostgreSQL-Verbindung, GORM-Setup. |
| `pkg/models/records.go` | **Datenmodelle**: GORM-Structs für Lebensmittel, Rezepte, Sonstiges. |
| `pkg/schema/schema.go` | **API-Schemas**: Request/Response-Structs für JSON-(De)Serialisierung. |
| `pkg/api/handler.go` | **Handler-Logik**: CRUD-Operationen für die API. |
| `pkg/mcp/server.go` | **MCP-Server**: KI-Tools (z. B. Rezeptvorschläge). |
| `pkg/utils/utils.go` | **Hilfsfunktionen**: ASCII-Banner, Helfer. |
| `pkg/config/config.go` | **Konfiguration**: Umgebungsvariablen, Standardwerte. |

---

## 🛠️ Werkzeuge & Befehle

### Go-Befehle
```bash
# Abhängigkeiten aktualisieren
go mod tidy

# Projekt bauen
go build ./cmd/api

# API-Server starten
go run ./cmd/api

# MCP-Server starten (geplant)
go run ./cmd/mcp

# Tests ausführen (falls vorhanden)
go test ./...
```

### Datenbank
```bash
# PostgreSQL-Container starten (falls lokal)
docker run --name kitchenops-db -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=kitchenops -p 5432:5432 -d postgres

# Verbindung testen
psql -h localhost -U postgres -d kitchenops
```

### Umgebungsvariablen
Erstelle eine `.env`-Datei in `pkg/config/`:
```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=kitchenops
POSTGRES_USERNAME=postgres
POSTGRES_PASSWORD=pass
POSTGRES_SCHEMA=public
```

---

## 📂 Projektstruktur (Erweitert)

```
kitchen-ops/
├── .vibe/
│   └── skills/
│       └── kitchen-ops/
│           └── SKILL.md          # Diese Datei
├── cmd/
│   ├── api/
│   │   └── main.go              # REST-API (Fiber)
│   └── mcp/
│       └── main.go              # MCP-Server (geplant)
├── pkg/
│   ├── config/
│   │   └── config.go            # Konfiguration (koanf)
│   ├── database/
│   │   └── db.go                # GORM + PostgreSQL
│   ├── models/
│   │   └── records.go           # GORM-Modelle
│   ├── repository/
│   │   └── record_repo.go       # Repository-Schicht (geplant)
│   ├── api/
│   │   └── handler.go           # HTTP-Handler
│   ├── mcp/
│   │   └── server.go            # MCP-Tools (geplant)
│   ├── schema/
│   │   └── schema.go            # Request/Response-Schemas
│   └── utils/
│       └── utils.go             # Hilfsfunktionen
├── go.mod
├── go.sum
└── README.md
```

---

## 🔍 Häufige Aufgaben

### 1. Neues Lebensmittel anlegen (API-Endpunkt)
**Schritte**:
1. Schema in `pkg/schema/schema.go` ergänzen:
   ```go
   type UpdateLebensmittelRecord struct {
       Name    string `json:"name"`
       Anzahl  uint16 `json:"anzahl"`
       Einheit string `json:"einheit"`
   }
   ```
2. Handler in `pkg/api/handler.go` hinzufügen:
   ```go
   func (h *DatabaseHandler) UpdateLebensmittel(ctx fiber.Ctx) error {
       id := ctx.Params("id")
       var input schema.UpdateLebensmittelRecord
       if err := ctx.Bind().Body(&input); err != nil {
           return fiber.NewError(fiber.StatusBadRequest, err.Error())
       }
       
       var lebensmittel models.Lebensmittel
       if err := h.db.First(&lebensmittel, id).Error; err != nil {
           return fiber.NewError(fiber.StatusNotFound, "Lebensmittel nicht gefunden")
       }
       
       lebensmittel.Name = input.Name
       lebensmittel.Anzahl = input.Anzahl
       lebensmittel.Einheit = input.Einheit
       
       if err := h.db.Save(&lebensmittel).Error; err != nil {
           return fiber.NewError(fiber.StatusInternalServerError, err.Error())
       }
       
       return ctx.JSON(schema.LebensmittelResponse{
           ID:      lebensmittel.ID,
           Name:    lebensmittel.Name,
           Anzahl:  lebensmittel.Anzahl,
           Einheit: lebensmittel.Einheit,
       })
   }
   ```
3. Route in `cmd/api/main.go` registrieren:
   ```go
   app.Put("/lebensmittel/:id", dbHandler.UpdateLebensmittel)
   ```

### 2. Rezept mit Zutaten erstellen
**Schritte**:
1. Modell in `pkg/models/records.go` prüfen (bereits vorhanden).
2. Schema in `pkg/schema/schema.go` ergänzen:
   ```go
   type NewRezeptRecord struct {
       Name                 string                `json:"name"`
       LebensmittelKategorie LebensmittelKategorie `json:"lebensmittel_kategorie"`
       EssensKategorie      EssensKategorie      `json:"essens_kategorie"`
       ZutatenIDs           []uint               `json:"zutaten_ids"`
   }
   ```
3. Handler implementieren (mit Many-to-Many-Beziehung):
   ```go
   func (h *DatabaseHandler) CreateRezept(ctx fiber.Ctx) error {
       var input schema.NewRezeptRecord
       if err := ctx.Bind().Body(&input); err != nil {
           return fiber.NewError(fiber.StatusBadRequest, err.Error())
       }
       
       // Zutaten laden
       var zutaten []models.Lebensmittel
       if err := h.db.Find(&zutaten, input.ZutatenIDs).Error; err != nil {
           return fiber.NewError(fiber.StatusNotFound, "Eine oder mehrere Zutaten nicht gefunden")
       }
       
       rezept := models.Rezept{
           Name:                 input.Name,
           LebensmittelKategorie: input.LebensmittelKategorie,
           EssensKategorie:      input.EssensKategorie,
           Zutaten:              zutaten,
       }
       
       if err := h.db.Create(&rezept).Error; err != nil {
           return fiber.NewError(fiber.StatusInternalServerError, err.Error())
       }
       
       return ctx.JSON(rezept)
   }
   ```

### 3. MCP-Tool für Rezeptvorschläge
**Schritte**:
1. Tool in `pkg/mcp/server.go` definieren:
   ```go
   package mcp
   
   import (
       "context"
       "github.com/mark3labs/mcp-go/server"
       "gitlab.com/ReneGumzej/kitchenops/pkg/models"
   )
   
   type KitchenOpsServer struct {
       db *gorm.DB
   }
   
   func (s *KitchenOpsServer) GetRezeptVorschlaege(ctx context.Context, req *schema.RezeptVorschlagRequest) (*schema.RezeptVorschlagResponse, error) {
       var rezepte []models.Rezept
       if err := s.db.Where("essens_kategorie = ?", req.Kategorie).Find(&rezepte).Error; err != nil {
           return nil, err
       }
       return &schema.RezeptVorschlagResponse{Rezepte: rezepte}, nil
   }
   ```
2. Server in `cmd/mcp/main.go` starten:
   ```go
   package main
   
   import (
       "gitlab.com/ReneGumzej/kitchenops/pkg/database"
       "gitlab.com/ReneGumzej/kitchenops/pkg/mcp"
   )
   
   func main() {
       // DB-Verbindung wie in cmd/api/main.go
       db, _ := database.DatabaseEngine(connectionString)
       
       mcpServer := mcp.NewKitchenOpsServer(db)
       server := mcp.NewServer(mcpServer)
       
       // Stdio oder SSE starten
       server.Start()
   }
   ```

---

## ⚠️ Wichtige Hinweise

1. **Datenbank-Sicherheit**:
   - **Nicht** `.env`-Dateien im Repository speichern (bereits in `.gitignore` prüfen).
   - **SSL-Modus**: Aktuell deaktiviert (`sslmode=disable`). Für Produktion auf `require` oder `verify-full` setzen.

2. **Fehlerbehandlung**:
   - **Immer** Fehler prüfen und zurückgeben (aktueller Code hat hier Lücken).
   - Nutze `fiber.Error` für HTTP-Fehler und `log.Fatal` für kritische Fehler.

3. **Performance**:
   - Für `GET /rezepte` mit vielen Ergebnissen: **Pagination** implementieren (z. B. mit `Limit`/`Offset` in GORM).
   - Indizes für häufig abgefragte Felder (`Name`, `Kategorie`) setzen.

4. **MCP-Integration**:
   - Tools sollten **idempotent** sein (gleiche Anfrage → gleiches Ergebnis).
   - **Dokumentation** der Tools in `SKILL.md` aktualisieren.

---

## 📚 Ressourcen

- [Fiber Dokumentation](https://docs.gofiber.io/)
- [GORM Dokumentation](https://gorm.io/docs/)
- [Koanf Dokumentation](https://github.com/knadh/koanf)
- [MCP-Spec](https://github.com/modelcontextprotocol/spec)
- [mcp-go Bibliothek](https://github.com/mark3labs/mcp-go)

---

## 🎯 Zusammenfassung

Diese `SKILL.md` definiert:
✅ **Code-Stil** (Deutsch, camelCase, PascalCase)
✅ **Architektur-Richtlinien** (Fiber, GORM, MCP)
✅ **Fehlerbehandlung** (korrekte HTTP-Statuscodes)
✅ **Werkzeuge & Befehle** (Go, PostgreSQL, Docker)
✅ **Häufige Aufgaben** (Neue Endpunkte, MCP-Tools)
✅ **Wichtige Hinweise** (Sicherheit, Performance)

**Ziel**: Konsistente und erweiterbare Entwicklung für **KitchenOps**.
