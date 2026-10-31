package models

import (
	"time"
)

type EssensKategorie string

const (
	EssensKategorieFruehstueck EssensKategorie = "Frühstück"
	EssensKategorieMittagessen EssensKategorie = "Mittagessen"
	EssensKategorieAbendessen  EssensKategorie = "Abendessen"
	EssensKategorieSnack       EssensKategorie = "Snack"
)

type LebensmittelKategorie string

const (
	LebensmittelKategoriePasta        LebensmittelKategorie = "Pasta"
	LebensmittelKategorieRindfleisch  LebensmittelKategorie = "Rindfleisch"
	LebensmittelKategorieHuehnerfleisch LebensmittelKategorie = "Hühnerfleisch"
	LebensmittelKategorieSchweinefleisch LebensmittelKategorie = "Schweinefleisch"
	LebensmittelKategorieVegetarisch  LebensmittelKategorie = "Vegetarisch"
	LebensmittelKategorieVegan        LebensmittelKategorie = "Vegan"
	LebensmittelKategorieSalat LebensmittelKategorie = "Salat"

)

type Makro struct {
	ID            uint    `gorm:"primaryKey;autoIncrement"`
	Protein       float32 `gorm:"type:numeric(6,2);not null"`
	Kohlenhydrate float32 `gorm:"type:numeric(6,2);not null"`
	Fett          float32 `gorm:"type:numeric(6,2);not null"`
}

// Lebensmittel repräsentiert eine Zutat inklusive der verknüpften Makros.
type Lebensmittel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(255);not null;index"`
	Anzahl    uint16    `gorm:"not null"`
	Einheit   string    `gorm:"type:varchar(50);not null"`
	// Küchenrichtung (einfache Zuordnung), nil = noch nicht zugeordnet
	LebensmittelKategorie *LebensmittelKategorie `gorm:"type:varchar(50)"`
	// Mahlzeiten (mehrere möglich), als JSON-Array gespeichert
	Mahlzeiten []EssensKategorie `gorm:"serializer:json;type:text"`
	// Gefrierfach markiert Lebensmittel, die in der Kühltruhe gelagert werden
	Gefrierfach bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// Sonstiges für Gegenstände oder Zutaten ohne Nährwertbezug.
type Sonstiges struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Anzahl    uint16    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Einkaufsartikel ist ein Eintrag auf der Einkaufsliste.
type Einkaufsartikel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(255);not null;index"`
	Anzahl    uint16    `gorm:"not null;default:1"`
	Einheit   string    `gorm:"type:varchar(50);not null"`
	Notiz     string    `gorm:"type:varchar(255)"`
	Erledigt  bool      `gorm:"not null;default:false"`
	Gefrierfach bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// RezeptZutatDetail speichert die Zutaten-Angaben eines Rezepts
// (inklusive Menge, Einheit und Inventar-Status) für die Anzeige.
type RezeptZutatDetail struct {
	Name      string `json:"name"`
	Menge     string `json:"menge"`
	Einheit   string `json:"einheit"`
	Vorhanden bool   `json:"vorhanden"`
}

// Rezept verbindet Kategorien und enthält eine Many-to-Many Beziehung zu Lebensmitteln.
type Rezept struct {
	ID                   uint                  `gorm:"primaryKey;autoIncrement"`
	Name                 string                `gorm:"type:varchar(255);not null"`
	LebensmittelKategorie LebensmittelKategorie `gorm:"type:varchar(50);not null"`
	EssensKategorie      EssensKategorie      `gorm:"type:varchar(50);not null"`
	Zutaten              []Lebensmittel        `gorm:"many2many:rezept_lebensmittel;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Vollständige Rezept-Details für die Anzeige
	Beschreibung        string              `gorm:"type:text"`
	ZutatenDetails      []RezeptZutatDetail `gorm:"serializer:json;type:text"`
	Schritte            []string            `gorm:"serializer:json;type:text"`
	ZubereitungszeitMin uint16              `gorm:"not null;default:0"`
	Schwierigkeit       string              `gorm:"type:varchar(20)"`
	Portionen           uint16              `gorm:"not null;default:0"`
	CreatedAt            time.Time           `gorm:"autoCreateTime"`
	UpdatedAt            time.Time           `gorm:"autoUpdateTime"`
}