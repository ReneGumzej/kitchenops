# Frontend App Prompt für KitchenOps

## Projektübersicht

Erstelle eine **moderne, benutzfreundliches Frontend** für das **KitchenOps-Backend** – ein Go-basiertes Küchenverwaltungssystem. Die App soll die REST-API des Backends nutzen und eine intuitive Oberfläche für die Verwaltung von Lebensmitteln, Rezepten und Kücheninventar bieten.

---

## 🎯 Backend-Integration

### API-Server
- **Base URL**: `http://<server-ip>:3000` (standardmäßig localhost:3000)
- **Framework**: Fiber (Go)
- **Datenformat**: JSON

### Verfügbare Endpunkte

| Methode | Endpunkt | Beschreibung | Request Body | Response |
|---------|----------|--------------|--------------|----------|
| GET | `/lebensmittel` | Alle Lebensmittel abrufen | - | `LebensmittelListResponse` |
| GET | `/lebensmittel/:id` | Einzelnes Lebensmittel | - | `LebensmittelResponse` |
| POST | `/lebensmittel` | Neues Lebensmittel erstellen | `NewLebensmittelRequest` | `LebensmittelResponse` |
| PUT | `/lebensmittel/:id` | Lebensmittel aktualisieren | `UpdateLebensmittelRequest` | `UpdateLebensmittelResponse` |
| DELETE | `/lebensmittel/:id` | Lebensmittel löschen | - | `DeleteLebensmittelResponse` |

### Datenstrukturen

**Lebensmittel**:
```dart
{
  "id": int,
  "name": "Mehl",
  "anzahl": 500,
  "einheit": "g",
  "created_at": "2026-09-28T10:00:00Z",
  "updated_at": "2026-09-28T10:00:00Z"
}
```

**Rezept** (geplant):
```dart
{
  "id": int,
  "name": "Spaghetti Bolognese",
  "lebensmittel_kategorie": "Pasta|Rindfleisch|Vegetarisch|Vegan|Salat|Hühnerfleisch|Schweinefleisch",
  "essens_kategorie": "Frühstück|Mittagessen|Abendessen|Snack",
  "zutaten": [Lebensmittel]
}
```

---

## 📱 App-Funktionen

### Kernfunktionen (MVP)
1. **Lebensmittelverwaltung**: CRUD-Operationen mit Suchfunktion und Filter
2. **Rezeptverwaltung**: Rezepte anzeigen, erstellen, bearbeiten mit Zutaten
3. **Inventarübersicht**: Bestandsüberblick mit Warnungen bei niedrigem Bestand
4. **Einkaufsliste**: Markieren, welche Lebensmittel nachgekauft werden müssen
5. **Suchfunktion**: Globale Suche über alle Lebensmittel und Rezepte

### Erweiterte Funktionen (Zukunft)
6. **Makro-Tracker**: Nährwertangaben für Lebensmittel
7. **Minküchen-Integration**: KI-gestützte Rezeptvorschläge (über MCP-Server)
8. **Barcode-Scanner**: Lebensmittel via Barcode scannen
9. **Ablaufdatum-Tracker**: Mindesthaltbarkeitsdatum erfassen
10. **Multi-User Support**: Gemeinsame Küchen für mehrere Nutzer

---

## 🎨 Design-System

### Farbpalette (KitchenOps Theme)

**Primärfarben** (Wärme & Küche):
- **Primary**: `#FF6B35` (Korallenorange)
- **Primary Light**: `#FF8E53`
- **Primary Dark**: `#E55A2B`
- **Secondary**: `#4ECDC4` (Türkis)
- **Secondary Light**: `#6ED3CF`
- **Secondary Dark**: `#3BB5AD`

**Neutrale Farben**:
- **Background**: `#FAFAFA`
- **Surface**: `#FFFFFF`
- **On Surface**: `#2D2D2D`
- **Grey**: `#9E9E9E`
- **Light Grey**: `#E0E0E0`

**Akzentfarben**:
- **Success**: `#4CAF50` (Grün)
- **Warning**: `#FFC107` (Gelb - niedriger Bestand)
- **Error**: `#F44336` (Rot)
- **Info**: `#2196F3` (Blau)

**Kategorie-Farben**:
- Frühstück: `#FFD700` (Gold)
- Mittagessen: `#FF6B35` (Orange)
- Abendessen: `#8E44AD` (Lila)
- Snack: `#4ECDC4` (Türkis)
- Vegetarisch: `#4CAF50` (Grün)
- Vegan: `#2E7D32` (Dunkelgrün)
- Pasta: `#FFB74D` (Pfirsich)
- Rindfleisch: `#D32F2F` (Rot)
- Hühnerfleisch: `#FFA000` (Amber)
- Schweinefleisch: `#C2185B` (Rosa)
- Salat: `#66BB6A` (Hellgrün)

### Typografie
- **Primary Font**: `Poppins` (Moderne Sans-Serif)
- **Secondary Font**: `Roboto` (Fallback)
- **Display Font**: `Bebas Neue` (für Überschriften)

**Schriftgrößen**:
- Headline 1: 32sp (App Bar)
- Headline 2: 28sp (Abschnitte)
- Headline 3: 24sp (Karten)
- Body Large: 18sp
- Body: 16sp
- Body Small: 14sp

### Design-Komponenten

**Cards**:
- Hintergrund: Surface (`#FFFFFF`)
- Elevation: 2 (standard), 8 (bei Fokus)
- Border Radius: 16
- Schatten: `BoxShadow(color: Colors.black12, blurRadius: 10, spreadRadius: 2)`
- Padding: 16 horizontal, 20 vertical

**Bottom Navigation Bar**:
- Hintergrund: Surface
- ausgewählter Tab: Primary Color
- nicht ausgewählter Tab: Grey
- Elevation: 8

---

### Routing
- `/` oder `/splash` - Splash Screen
- `/onboarding` - Onboarding
- `/home` - Hauptbildschirm
- `/lebensmittel` - Lebensmittel-Liste
- `/lebensmittel/:id` - Lebensmittel-Detail
- `/lebensmittel/add` - Lebensmittel hinzufügen
- `/rezepte` - Rezept-Liste
- `/inventar` - Inventar
- `/einkaufsliste` - Einkaufsliste

## 🚀 Fazit

Erstelle eine **moderne Frontend-App** mit:
- ✅ **Vollständiger Backend-Integration** (alle API-Endpunkte)
- ✅ **KitchenOps-Design-System** (Farben, Typografie, Komponenten)
- ✅ **Intuitiver UI/UX** (Material Design 3, flüssige Animationen)
- ✅ **Robuster Architektur**
- ✅ **Offline-Fähigkeit** (basic Caching)

Die App soll **Küchen-Enthusiasten** helfen, ihre Vorräte zu verwalten und den Überblick zu behalten.

**Start jetzt und baue die beste Küchen-App!** 🍳
