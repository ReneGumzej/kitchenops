// Typen der LLM-Antworten. Die jsonschema-Tags werden vom GoAI-Framework
// per Reflexion in JSON-Schemas übersetzt (structured output) - die Felder
// sind damit Pflicht und Enums werden erzwungen.
package llm

// Anfrage ist die strukturierte Anfrage für die Rezept-Generierung.
type Anfrage struct {
	Prompt                string // freier Text, z.B. "bitte zwei Gerichte"
	EssensKategorie       string // Frühstück | Mittagessen | Abendessen | Snack
	LebensmittelKategorie string // Pasta | Rindfleisch | ... | Vegan
	Anzahl                uint   // gewünschte Anzahl an Rezepten
}

// RezeptZutat ist eine Zutat aus der LLM-Antwort.
type RezeptZutat struct {
	Name      string `json:"name" jsonschema:"description=Name der Zutat"`
	Menge     string `json:"menge" jsonschema:"description=z.B. 200 oder 0.5"`
	Einheit   string `json:"einheit" jsonschema:"description=z.B. g, ml, Stück"`
	Vorhanden bool   `json:"vorhanden" jsonschema:"description=true, wenn die Zutat im Inventar ist"`
}

// Rezept ist ein vollständiger Rezeptvorschlag aus der LLM-Antwort.
type Rezept struct {
	Name                  string        `json:"name" jsonschema:"description=Name des Gerichts"`
	Beschreibung          string        `json:"beschreibung" jsonschema:"description=Kurze, appetitliche Beschreibung (2-3 Sätze)"`
	EssensKategorie       string        `json:"essens_kategorie" jsonschema:"enum=Frühstück|Mittagessen|Abendessen|Snack"`
	LebensmittelKategorie string        `json:"lebensmittel_kategorie" jsonschema:"enum=Pasta|Rindfleisch|Hühnerfleisch|Schweinefleisch|Vegetarisch|Vegan|Salat"`
	Zutaten               []RezeptZutat `json:"zutaten" jsonschema:"description=Zutaten mit Menge und Einheit"`
	Schritte              []string      `json:"schritte" jsonschema:"description=Zubereitungsschritte in der richtigen Reihenfolge"`
	ZubereitungszeitMin   uint16        `json:"zubereitungszeit_minuten" jsonschema:"description=Gesamtdauer in Minuten"`
	Schwierigkeit         string        `json:"schwierigkeit" jsonschema:"enum=Einfach|Mittel|Schwer"`
	Portionen             uint16        `json:"portionen"`
}

// RezeptAntwort ist die Hüllstruktur der LLM-Antwort.
type RezeptAntwort struct {
	Rezepte []Rezept `json:"rezepte"`
}

// KlassifiziereAnfrageEintrag ist ein zu klassifizierendes Lebensmittel.
type KlassifiziereAnfrageEintrag struct {
	ID   uint
	Name string
}

// KlassifiziereEintrag ist ein Eintrag der Klassifizierungs-Antwort.
type KlassifiziereEintrag struct {
	ID                    uint     `json:"id" jsonschema:"description=ID des Lebensmittels aus der Liste"`
	LebensmittelKategorie string   `json:"lebensmittel_kategorie" jsonschema:"enum=|Pasta|Rindfleisch|Hühnerfleisch|Schweinefleisch|Vegetarisch|Vegan|Salat,description=Küchenrichtung, leer wenn unklar"`
	Mahlzeiten            []string `json:"mahlzeiten" jsonschema:"description=Passende Mahlzeiten"`
}

// KlassifiziereAntwort ist die Hüllstruktur der Klassifizierungs-Antwort.
type KlassifiziereAntwort struct {
	Zuordnungen []KlassifiziereEintrag `json:"zuordnungen"`
}

// ErfassungEintrag ist ein aus Freitext extrahiertes Lebensmittel.
type ErfassungEintrag struct {
	Name        string `json:"name" jsonschema:"description=Name des Lebensmittels, bereinigt (ohne Mengenangaben)"`
	Anzahl      uint16 `json:"anzahl" jsonschema:"description=Menge; ohne Angabe im Text: 1"`
	Einheit     string `json:"einheit" jsonschema:"enum=Stück|g|kg|ml|l|Packung|Beutel|Dose|Glas|Flasche|Becher|Bund|Tafel,description=Einheit; ohne Angabe im Text: Stück"`
	Gefrierfach bool   `json:"gefrierfach" jsonschema:"description=true, wenn das Lebensmittel ins Gefrierfach soll"`
}

// ErfassungAntwort ist die Hüllstruktur der Erfassungs-Antwort.
type ErfassungAntwort struct {
	Lebensmittel []ErfassungEintrag `json:"lebensmittel"`
}
