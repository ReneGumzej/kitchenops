# KitchenOps - Frontend

Moderne Multi-Page-Web-App (HTML, CSS, vanilla JavaScript) für die Küchenverwaltung.
Jede Ansicht ist eine eigene HTML-Seite mit eigenem JavaScript-Modul.

## 📁 Projektstruktur

```
frontend/
├── index.html           # Dashboard (Startseite)
├── lebensmittel.html    # Lebensmittel-Verwaltung (CRUD)
├── einkaufsliste.html   # Einkaufsliste
├── inventar.html        # Inventar (Kategorien-Ansicht)
├── einstellungen.html   # Einstellungen (Server, Theme, Limits)
├── css/
│   ├── style.css        # Design-System (Basis)
│   └── pages.css        # Seitenspezifische Komponenten
├── js/
│   ├── api.js           # REST-Client (Fetch)
│   ├── common.js        # Geteilte Logik: Store, Toast, Modal, Layout, DataManager
│   ├── dashboard.js     # Dashboard-Logik
│   ├── lebensmittel.js  # Lebensmittel-Logik
│   ├── einkaufsliste.js # Einkaufslisten-Logik
│   ├── inventar.js      # Inventar-Logik
│   └── einstellungen.js # Einstellungs-Logik
├── assets/             # Bilder, Icons (optional)
├── nginx.conf           # NGINX-Konfiguration
├── docker-compose.yml  # Docker-Komposition
└── README.md
```

## 📄 Seiten & Funktionen

| Seite | Funktionen |
|-------|-----------|
| **index.html** | Statistiken, Schnellaktionen, neueste Lebensmittel, Verbindungsstatus |
| **lebensmittel.html** | Liste (Grid/Tabelle), Suche, Filter (alle/niedrig/kritisch/Einkaufsliste), Anlegen, Bearbeiten, Löschen, Detailansicht via `?id=` |
| **einkaufsliste.html** | Einkaufsliste im Backend persistiert (SQLite): Artikel hinzufügen, abhaken, löschen, Liste leeren, Nachkauf-Vorschläge aus kritischem Bestand |
| **inventar.html** | Gruppierung nach Kategorien (Trockenware, Flüssiges, Konserven, Stückware), Kategorie-Filter, Statistik je Gruppe |
| **einstellungen.html** | Backend-URL speichern/testen, Verbindungsstatus, Theme (Hell/Dunkel/System), Bestands-Schwellenwerte, lokale Daten verwalten |

---

## 🚀 Schnellstart

### Option 1: Direkte Nutzung (für Entwicklung)

1. **Backend starten** (falls noch nicht läuft):
   ```bash
   cd /home/reneg/Projekte/local_dev/go/kitchen-ops
   go run main.go
   ```

2. **Frontend öffnen**:
   - Einfach die `index.html` in einem Browser öffnen
   - Oder mit Python HTTP Server:
     ```bash
     cd frontend
     python3 -m http.server 8000
     ```
   - Dann im Browser: `http://localhost:8000` öffnen

### Option 2: Mit Docker (für Produktion)

1. **Docker Container bauen und starten**:
   ```bash
   cd /home/reneg/Projekte/local_dev/go/kitchen-ops
   docker-compose up -d
   ```

2. **Zugriff**:
   - Frontend: `http://localhost:8080`
   - Backend API: `http://localhost:3000`

### Option 3: Nur Frontend mit Docker

```bash
cd /home/reneg/Projekte/local_dev/go/kitchen-ops/frontend
docker build -t kitchenops-frontend .
docker run -p 8080:80 kitchenops-frontend
```

---

## ✨ Features

### 🎯 Lebensmittelverwaltung
- **Liste aller Lebensmittel** mit Suchfunktion
- **Neues Lebensmittel hinzufügen**
- **Lebensmittel bearbeiten**
- **Lebensmittel löschen** (mit Bestätigung)
- **Detailansicht** für jedes Lebensmittel

### 📊 Dashboard
- Übersicht über alle Lebensmittel
- Statistiken (Gesamtanzahl, niedriger Bestand, Einkaufsliste)
- Schnellaktionen
- Vorschau der letzten Lebensmittel

### 🛒 Einkaufsliste
- Automatische Erkennung von Lebensmitteln mit niedrigem Bestand
- Manuelles Markieren für Einkaufsliste
- "Alle als gekauft markieren" Funktion
- Übersichtlicher Status

### 📋 Inventar
- Gruppierung nach Kategorien
- Übersicht über alle Vorräte

### ⚙️ Einstellungen
- **Server-URL Konfiguration** (für Verbindung zum Backend)
- **Design-Umschaltung** (Hell/Dunkel)
- **Verbindungsstatus** Anzeige

---

## 🎨 Design-System

### Farben (KitchenOps Theme)
- **Primary**: `#FF6B35` (Korallenorange)
- **Secondary**: `#4ECDC4` (Türkis)
- **Success**: `#4CAF50` (Grün)
- **Warning**: `#FFC107` (Gelb)
- **Error**: `#F44336` (Rot)
- **Background**: `#FAFAFA` (Hell) / `#121212` (Dunkel)
- **Surface**: `#FFFFFF` (Hell) / `#1E1E1E` (Dunkel)

### Typografie
- **Primary Font**: Poppins (von Google Fonts)
- **Secondary Font**: Roboto (von Google Fonts)
- **Display Font**: Bebas Neue (von Google Fonts)

### Komponenten
- **Cards**: Abgerundete Ecken (16px), Schatten, Hover-Effekte
- **Buttons**: Mehrere Varianten (Primary, Secondary, Outline)
- **Formulare**: Moderne Input-Felder mit Validation
- **Toast Notifications**: Benachrichtigungen für Aktionen
- **Modal Dialogs**: Für Formulare und Details

---

## 🔧 Backend-Integration

### API Endpunkte

| Methode | Endpunkt | Beschreibung |
|---------|----------|--------------|
| GET | `/lebensmittel` | Alle Lebensmittel abrufen |
| GET | `/lebensmittel/:id` | Einzelnes Lebensmittel abrufen |
| POST | `/lebensmittel` | Neues Lebensmittel erstellen |
| PUT | `/lebensmittel/:id` | Lebensmittel aktualisieren |
| DELETE | `/lebensmittel/:id` | Lebensmittel löschen |

### Datenstruktur

**Lebensmittel-Objekt:**
```json
{
  "id": 1,
  "name": "Mehl",
  "anzahl": 500,
  "einheit": "g",
  "created_at": "2026-10-31T10:00:00Z",
  "updated_at": "2026-10-31T10:00:00Z"
}
```

**Neues Lebensmittel (Request):**
```json
{
  "name": "Mehl",
  "anzahl": 500,
  "einheit": "g"
}
```

---

## 📱 Responsive Design

Die App ist voll responsiv und funktioniert auf:
- **Desktop** (ab 1200px)
- **Tablet** (768px - 1199px)
- **Mobile** (unter 768px)

### Mobile-spezifische Features
- **Hamburger-Menü** für Navigation
- **Slide-in Sidebar**
- **Touch-optimierte Buttons**
- **Angepasste Grid-Layouts**

---

## 🔌 Anpassungen

### Server-URL ändern

1. Im Browser die App öffnen
2. Zu **Einstellungen** navigieren
3. Neue Server-URL eingeben (z. B. `http://your-server:3000`)
4. Auf **Testen** klicken

### Standard-Einheiten hinzufügen

In `app.js` die `COMMON_UNITS` Konstante anpassen:
```javascript
const CONSTANTS = {
    COMMON_UNITS: ['g', 'kg', 'mg', 'l', 'ml', 'Stück', 'Deine Einheit'],
    // ...
};
```

### Stock Level Schwellenwerte ändern

In `app.js` die Konstanten anpassen:
```javascript
const CONSTANTS = {
    LOW_STOCK_THRESHOLD: 5,    // Unter diesem Wert = "Niedrig"
    CRITICAL_STOCK_THRESHOLD: 2,  // Unter diesem Wert = "Kritisch"
    // ...
};
```

---

## 🛠️ Browser Support

- ✅ Chrome (empfohlen)
- ✅ Firefox
- ✅ Safari
- ✅ Edge
- ⚠️ Internet Explorer (nicht unterstützt)

---

## 📦 Dependencies

Die App verwendet nur:
- **Google Fonts** (Poppins, Roboto, Bebas Neue)
- **Font Awesome 6** (Icons)
- **Keine npm/Package.json nötig** - reine HTML/CSS/JS

---

## 💡 Tipps

### Entwicklung
- Nutze die Browser DevTools (F12) für Debugging
- Die App nutzt **LocalStorage** für Theme-Einstellungen
- **CORS** muss auf dem Backend aktiviert sein

### Produktion
- Für bessere Performance: CSS/JS minifizieren
- Bilder optimieren
- CDN für Font Awesome und Google Fonts nutzen

### Backend
- Stelle sicher, dass CORS auf dem Backend aktiviert ist:
  ```go
  // In deinem Go-Backend (main.go oder Router)
  app.Use(cors.New(cors.Config{
      AllowOrigins:     "*",
      AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
      AllowHeaders:     "Origin, Content-Type, Accept",
      ExposeHeaders:    "Content-Length",
      AllowCredentials: true,
  }))
  ```

---

## 📄 Lizenz

Dieses Projekt ist Teil von KitchenOps und kann frei verwendet werden.

---

## 🙏 Unterstützung

Bei Problemen oder Fragen:
- Backend-Logs prüfen
- Browser Console (F12) öffnen
- Netzwerk-Tab prüfen (ob Requests erfolgreich sind)

---

**Viel Spaß mit KitchenOps!** 🍳
