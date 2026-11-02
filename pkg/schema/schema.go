package schema

import "time"

type LebensmittelResponse struct {
	ID uint 			`json:"id"`
	Name string			`json:"name"`
	Anzahl uint16		`json:"anzahl"`
	Einheit string		`json:"einheit"`
	LebensmittelKategorie string `json:"lebensmittel_kategorie"`
	Mahlzeiten []string `json:"mahlzeiten"`
	Gefrierfach bool `json:"gefrierfach"`
}

// LebensmittelListResponse is a wrapper for returning a list of LebensmittelResponse
type LebensmittelListResponse struct {
	Lebensmittel []LebensmittelResponse `json:"lebensmittel"`
}

type NewLebensmittelRequest struct {
	Name string			`json:"name"`
	Anzahl uint16		`json:"anzahl"`
	Einheit string		`json:"einheit"`
	// Optional: Kategorien. Ohne Angabe wird der Regel-Klassifikator genutzt.
	LebensmittelKategorie string `json:"lebensmittel_kategorie"`
	Mahlzeiten []string `json:"mahlzeiten"`
	Gefrierfach bool `json:"gefrierfach"`
}

type UpdateLebensmittelRequest struct{
	Name string			`json:"name"`
	Anzahl uint16		`json:"anzahl"`
	Einheit string		`json:"einheit"`
	LebensmittelKategorie string `json:"lebensmittel_kategorie"`
	Mahlzeiten []string `json:"mahlzeiten"`
	// Gefrierfach als Pointer: nur setzen, wenn das Feld im Request
	// tatsächlich ankommt (schützt vor Clients, die es nicht senden).
	Gefrierfach *bool `json:"gefrierfach"`
}

type UpdateLebensmittelResponse struct {
	ID uint 			`json:"id"`
	Name string			`json:"name"`
	Anzahl uint16		`json:"anzahl"`
	Einheit string		`json:"einheit"`
}


type DeleteLebensmittelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KategorisierenResponse fasst das Ergebnis der automatischen
// Kategorisierung zusammen.
type KategorisierenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PerRegeln int `json:"per_regeln"`
	PerLLM int `json:"per_llm"`
	Offen int `json:"offen"`
}

// LebensmittelEingabeRequest ist die Freitext-Eingabe für die
// KI-gestützte Lebensmittelerfassung.
type LebensmittelEingabeRequest struct {
	Text string `json:"text"`
}

// EinkaufsartikelResponse ist die Antwort für einen einzelnen Einkaufslisten-Eintrag.
// InventarHinweis beschreibt, was beim Abhaken automatisch ins Inventar
// übernommen wurde (leer, wenn kein Übernahmevorgang stattfand).
type EinkaufsartikelResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Anzahl          uint16 `json:"anzahl"`
	Einheit         string `json:"einheit"`
	Notiz           string `json:"notiz"`
	Erledigt        bool   `json:"erledigt"`
	Gefrierfach     bool   `json:"gefrierfach"`
	InventarHinweis string `json:"inventar_hinweis"`
}

// EinkaufsartikelListResponse ist ein Wrapper für eine Liste von Einkaufslisten-Einträgen.
type EinkaufsartikelListResponse struct {
	Artikel []EinkaufsartikelResponse `json:"artikel"`
}

// NewEinkaufsartikelRequest zum Anlegen eines neuen Einkaufslisten-Eintrags.
type NewEinkaufsartikelRequest struct {
	Name        string `json:"name"`
	Anzahl      uint16 `json:"anzahl"`
	Einheit     string `json:"einheit"`
	Notiz       string `json:"notiz"`
	Gefrierfach bool   `json:"gefrierfach"`
}

// UpdateEinkaufsartikelRequest zum Bearbeiten eines Einkaufslisten-Eintrags.
type UpdateEinkaufsartikelRequest struct {
	Name        string `json:"name"`
	Anzahl      uint16 `json:"anzahl"`
	Einheit     string `json:"einheit"`
	Notiz       string `json:"notiz"`
	Erledigt    bool   `json:"erledigt"`
	Gefrierfach *bool  `json:"gefrierfach"`
}

// DeleteEinkaufsartikelResponse bestätigt das Löschen eines Einkaufslisten-Eintrags.
type DeleteEinkaufsartikelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type EinkaufzettelEingabeRequest struct {
	Text string `json:"text"`
}

// =============================================================================
// Rezepte
// =============================================================================

// RezeptGenerierenRequest für die KI-Rezept-Generierung.
// Prompt, EssensKategorie und LebensmittelKategorie sind optional kombinierbar.
type RezeptGenerierenRequest struct {
	Prompt                string `json:"prompt"`
	EssensKategorie       string `json:"essens_kategorie"`
	LebensmittelKategorie string `json:"lebensmittel_kategorie"`
	Anzahl                uint   `json:"anzahl"`
}

// RezeptZutatResponse ist eine Zutat eines (generierten) Rezepts.
type RezeptZutatResponse struct {
	Name      string `json:"name"`
	Menge     string `json:"menge"`
	Einheit   string `json:"einheit"`
	Vorhanden bool   `json:"vorhanden"`
}

// RezeptResponse ist ein vollständiges Rezept (generiert oder gespeichert).
type RezeptResponse struct {
	Name                  string                `json:"name"`
	Beschreibung          string                `json:"beschreibung"`
	EssensKategorie       string                `json:"essens_kategorie"`
	LebensmittelKategorie string                `json:"lebensmittel_kategorie"`
	Zutaten               []RezeptZutatResponse `json:"zutaten"`
	Schritte              []string              `json:"schritte"`
	ZubereitungszeitMin   uint16                `json:"zubereitungszeit_minuten"`
	Schwierigkeit         string                `json:"schwierigkeit"`
	Portionen             uint16                `json:"portionen"`
}

// RezeptListResponse ist die Antwort der Rezept-Generierung.
type RezeptListResponse struct {
	Rezepte []RezeptResponse `json:"rezepte"`
}

// RezeptZutatRequest ist eine Zutat eines zu speichernden Rezepts
// mit vollständigen Mengenangaben.
type RezeptZutatRequest struct {
	Name      string `json:"name"`
	Menge     string `json:"menge"`
	Einheit   string `json:"einheit"`
	Vorhanden bool   `json:"vorhanden"`
}

// SaveRezeptRequest speichert ein Rezept mit vollständigen Details
// und Verknüpfung auf Inventar-IDs.
type SaveRezeptRequest struct {
	Name                  string              `json:"name"`
	LebensmittelKategorie string              `json:"lebensmittel_kategorie"`
	EssensKategorie       string              `json:"essens_kategorie"`
	Beschreibung          string              `json:"beschreibung"`
	Zutaten               []RezeptZutatRequest `json:"zutaten"`
	Schritte              []string            `json:"schritte"`
	ZubereitungszeitMin   uint16              `json:"zubereitungszeit_minuten"`
	Schwierigkeit         string              `json:"schwierigkeit"`
	Portionen             uint16              `json:"portionen"`
	ZutatenIDs            []uint              `json:"zutaten_ids"`
}

// GespeichertesRezeptResponse ist ein gespeichertes Rezept aus der DB.
// Die Feldnamen entsprechen bewusst der RezeptResponse der KI-Generierung,
// damit das Frontend beide Formate identisch rendern kann.
type GespeichertesRezeptResponse struct {
	ID                    uint                 `json:"id"`
	Name                  string               `json:"name"`
	Beschreibung          string               `json:"beschreibung"`
	EssensKategorie       string               `json:"essens_kategorie"`
	LebensmittelKategorie string               `json:"lebensmittel_kategorie"`
	Zutaten               []RezeptZutatResponse `json:"zutaten"`
	Schritte              []string             `json:"schritte"`
	ZubereitungszeitMin   uint16               `json:"zubereitungszeit_minuten"`
	Schwierigkeit         string               `json:"schwierigkeit"`
	Portionen             uint16               `json:"portionen"`
	CreatedAt             time.Time            `json:"created_at"`
}

// GespeichertesRezeptListResponse ist die Liste aller gespeicherten Rezepte.
type GespeichertesRezeptListResponse struct {
	Rezepte []GespeichertesRezeptResponse `json:"rezepte"`
}

// SaveRezeptResponse bestätigt das Speichern eines Rezepts.
type SaveRezeptResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      uint   `json:"id"`
}

// DeleteRezeptResponse bestätigt das Löschen eines Rezepts.
type DeleteRezeptResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}