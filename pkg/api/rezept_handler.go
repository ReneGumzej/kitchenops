package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"gitlab.com/ReneGumzej/kitchenops/pkg/llm"
	"gitlab.com/ReneGumzej/kitchenops/pkg/logging"
	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
	"gitlab.com/ReneGumzej/kitchenops/pkg/schema"
)

// RezeptHandler verbindet Datenbank und LLM-Client für die Rezept-Funktionen.
type RezeptHandler struct {
	db  *gorm.DB
	llm *llm.Client
}

func NewRezeptHandler(db *gorm.DB, client *llm.Client) *RezeptHandler {
	return &RezeptHandler{
		db:  db,
		llm: client,
	}
}

// gültige Werte für die strukturierte Vorauswahl
var validEssensKategorien = map[string]bool{
	string(models.EssensKategorieFruehstueck): true,
	string(models.EssensKategorieMittagessen): true,
	string(models.EssensKategorieAbendessen):  true,
	string(models.EssensKategorieSnack):       true,
}

var validLebensmittelKategorien = map[string]bool{
	string(models.LebensmittelKategoriePasta):         true,
	string(models.LebensmittelKategorieRindfleisch):   true,
	string(models.LebensmittelKategorieHuehnerfleisch): true,
	string(models.LebensmittelKategorieSchweinefleisch): true,
	string(models.LebensmittelKategorieVegetarisch):   true,
	string(models.LebensmittelKategorieVegan):         true,
	string(models.LebensmittelKategorieSalat):        true,
}

// GenerateRezepte (POST /rezepte/generieren) lädt das Inventar, sendet es
// zusammen mit der strukturierten Vorauswahl und/oder dem freien Prompt
// an das LLM und gibt die passenden Rezeptvorschläge strukturiert zurück.
func (handler *RezeptHandler) GenerateRezepte(ctx fiber.Ctx) error {
	request := &schema.RezeptGenerierenRequest{}

	if err := ctx.Bind().Body(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}

	// Validierung der strukturierten Vorauswahl (falls angegeben)
	if request.EssensKategorie != "" && !validEssensKategorien[request.EssensKategorie] {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Essenskategorie: "+request.EssensKategorie)
	}
	if request.LebensmittelKategorie != "" && !validLebensmittelKategorien[request.LebensmittelKategorie] {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Lebensmittelkategorie: "+request.LebensmittelKategorie)
	}

	// Es muss mindestens eine Angabe gemacht werden
	if request.Prompt == "" && request.EssensKategorie == "" && request.LebensmittelKategorie == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Bitte geben Sie einen Wunsch-Text oder eine Kategorie-Auswahl an")
	}

	if !handler.llm.Verfügbar() {
		return fiber.NewError(fiber.StatusServiceUnavailable,
			"Rezept-Generierung nicht konfiguriert: bitte einen API-Key für Ihren LLM-Anbieter als Umgebungsvariable setzen (z.B. MISTRAL_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY)")
	}

	// Inventar laden
	var inventar []models.Lebensmittel
	if err := handler.db.Find(&inventar).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Abrufen des Inventars: "+err.Error())
	}

	// Bei vegetarischer/veganer Vorauswahl: Fleisch vorab aus dem
	// Inventar filtern (kürzere Prompts, bessere Trefferquote)
	if request.LebensmittelKategorie == string(models.LebensmittelKategorieVegetarisch) ||
		request.LebensmittelKategorie == string(models.LebensmittelKategorieVegan) {
		gefiltert := inventar[:0]
		for _, lm := range inventar {
			if lm.LebensmittelKategorie != nil && lm.LebensmittelKategorie.IstFleisch() {
				continue
			}
			if request.LebensmittelKategorie == string(models.LebensmittelKategorieVegan) &&
				lm.LebensmittelKategorie != nil &&
				*lm.LebensmittelKategorie == models.LebensmittelKategorieVegetarisch {
				// Milchprodukte/Eier sind nicht vegan
				continue
			}
			gefiltert = append(gefiltert, lm)
		}
		inventar = gefiltert
	}

	// LLM aufrufen
	start := time.Now()
	rezepte, err := handler.llm.GeneriereRezepte(inventar, llm.Anfrage{
		Prompt:                request.Prompt,
		EssensKategorie:       request.EssensKategorie,
		LebensmittelKategorie: request.LebensmittelKategorie,
		Anzahl:                request.Anzahl,
	})
	if err != nil {
		logging.Error("rezept-generierung fehlgeschlagen",
			"essens_kategorie", request.EssensKategorie,
			"lebensmittel_kategorie", request.LebensmittelKategorie,
			"fehler", err.Error())
		return fiber.NewError(fiber.StatusBadGateway, "Fehler bei der Rezept-Generierung: "+err.Error())
	}
	logging.Info("rezepte generiert",
		"essens_kategorie", request.EssensKategorie,
		"lebensmittel_kategorie", request.LebensmittelKategorie,
		"inventar", len(inventar),
		"ergebnisse", len(rezepte),
		"dauer_ms", time.Since(start).Milliseconds(),
	)

	// LLM-Ergebnis in API-Schema überführen
	responses := make([]schema.RezeptResponse, len(rezepte))
	for i, r := range rezepte {
		zutaten := make([]schema.RezeptZutatResponse, len(r.Zutaten))
		for j, z := range r.Zutaten {
			zutaten[j] = schema.RezeptZutatResponse{
				Name:      z.Name,
				Menge:     z.Menge,
				Einheit:   z.Einheit,
				Vorhanden: z.Vorhanden,
			}
		}
		responses[i] = schema.RezeptResponse{
			Name:                  r.Name,
			Beschreibung:          r.Beschreibung,
			EssensKategorie:       r.EssensKategorie,
			LebensmittelKategorie: r.LebensmittelKategorie,
			Zutaten:               zutaten,
			Schritte:              r.Schritte,
			ZubereitungszeitMin:   r.ZubereitungszeitMin,
			Schwierigkeit:         r.Schwierigkeit,
			Portionen:             r.Portionen,
		}
	}

	return ctx.JSON(schema.RezeptListResponse{Rezepte: responses})
}

// SaveRezept (POST /rezepte) speichert ein Rezept in der DB und verknüpft
// die Zutaten über die übergebenen Lebensmittel-IDs (Many-to-Many).
func (handler *RezeptHandler) SaveRezept(ctx fiber.Ctx) error {
	request := &schema.SaveRezeptRequest{}

	if err := ctx.Bind().Body(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}

	if request.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name ist ein Pflichtfeld")
	}
	if !validEssensKategorien[request.EssensKategorie] {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Essenskategorie: "+request.EssensKategorie)
	}
	if !validLebensmittelKategorien[request.LebensmittelKategorie] {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Lebensmittelkategorie: "+request.LebensmittelKategorie)
	}

	// Existierende Lebensmittel zu den übergebenen IDs laden
	zutaten := []models.Lebensmittel{}
	if len(request.ZutatenIDs) > 0 {
		if err := handler.db.Find(&zutaten, request.ZutatenIDs).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Laden der Zutaten: "+err.Error())
		}
	}

	// Zutaten-Details für die spätere Anzeige übernehmen
	zutatenDetails := make([]models.RezeptZutatDetail, 0, len(request.Zutaten))
	for _, z := range request.Zutaten {
		if z.Name == "" {
			continue
		}
		zutatenDetails = append(zutatenDetails, models.RezeptZutatDetail{
			Name:      z.Name,
			Menge:     z.Menge,
			Einheit:   z.Einheit,
			Vorhanden: z.Vorhanden,
		})
	}

	rezept := models.Rezept{
		Name:                  request.Name,
		LebensmittelKategorie: models.LebensmittelKategorie(request.LebensmittelKategorie),
		EssensKategorie:       models.EssensKategorie(request.EssensKategorie),
		Zutaten:               zutaten,
		Beschreibung:          request.Beschreibung,
		ZutatenDetails:        zutatenDetails,
		Schritte:              request.Schritte,
		ZubereitungszeitMin:   request.ZubereitungszeitMin,
		Schwierigkeit:         request.Schwierigkeit,
		Portionen:             request.Portionen,
	}
	if err := handler.db.Create(&rezept).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Speichern des Rezepts: "+err.Error())
	}

	logging.Info("rezept gespeichert",
		"id", rezept.ID, "name", rezept.Name,
		"lebensmittel_kategorie", string(rezept.LebensmittelKategorie),
		"essens_kategorie", string(rezept.EssensKategorie),
		"zutaten_verknüpft", len(zutaten),
		"zutaten_details", len(zutatenDetails),
	)

	return ctx.JSON(schema.SaveRezeptResponse{
		Success: true,
		Message: "Rezept erfolgreich gespeichert",
		ID:      rezept.ID,
	})
}

// GetAllRezepte (GET /rezepte) liefert alle gespeicherten Rezepte inkl. Zutaten.
func (handler *RezeptHandler) GetAllRezepte(ctx fiber.Ctx) error {
	var rezepte []models.Rezept

	if err := handler.db.Preload("Zutaten").Find(&rezepte).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Abrufen der Rezepte: "+err.Error())
	}

	responses := make([]schema.GespeichertesRezeptResponse, len(rezepte))
	for i, r := range rezepte {
		// Zutaten-Details verwenden (vollständig); Fallback auf die
		// verknüpften Lebensmittelnamen für alte Einträge ohne Details
		zutaten := make([]schema.RezeptZutatResponse, 0, len(r.ZutatenDetails))
		for _, zd := range r.ZutatenDetails {
			zutaten = append(zutaten, schema.RezeptZutatResponse{
				Name:      zd.Name,
				Menge:     zd.Menge,
				Einheit:   zd.Einheit,
				Vorhanden: zd.Vorhanden,
			})
		}
		if len(zutaten) == 0 {
			for _, z := range r.Zutaten {
				zutaten = append(zutaten, schema.RezeptZutatResponse{
					Name:      z.Name,
					Menge:     "",
					Einheit:   z.Einheit,
					Vorhanden: true,
				})
			}
		}

		responses[i] = schema.GespeichertesRezeptResponse{
			ID:                    r.ID,
			Name:                  r.Name,
			Beschreibung:          r.Beschreibung,
			EssensKategorie:       string(r.EssensKategorie),
			LebensmittelKategorie: string(r.LebensmittelKategorie),
			Zutaten:               zutaten,
			Schritte:              r.Schritte,
			ZubereitungszeitMin:   r.ZubereitungszeitMin,
			Schwierigkeit:         r.Schwierigkeit,
			Portionen:             r.Portionen,
			CreatedAt:             r.CreatedAt,
		}
	}

	return ctx.JSON(schema.GespeichertesRezeptListResponse{Rezepte: responses})
}

// ErfasseLebensmittelPerText (POST /lebensmittel/erfassen) liest einen
// freien Text (z.B. "Füge mir Brot, Butter und zwei mal Sahne hinzu und
// packe das Hähnchen ins Gefrierfach"), lässt ihn vom LLM strukturieren
// und trägt die Lebensmittel in die Datenbank ein:
// - neue Artikel werden mit Regel-Kategorien angelegt (Gefrierfach-Flag
//   aus dem Text)
// - bereits vorhandene Artikel erhalten die Menge aufaddiert (mit
//   Einheiten-Umrechnung); ein Gefrierfach-Hinweis setzt das Flag.
func (handler *RezeptHandler) ErfasseLebensmittelPerText(ctx fiber.Ctx) error {
	request := &schema.LebensmittelEingabeRequest{}

	if err := ctx.Bind().Body(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}
	if request.Text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Bitte geben Sie einen Text ein")
	}
	if !handler.llm.Verfügbar() {
		return fiber.NewError(fiber.StatusServiceUnavailable,
			"Text-Erfassung nicht konfiguriert: bitte einen API-Key für Ihren LLM-Anbieter als Umgebungsvariable setzen (z.B. MISTRAL_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY)")
	}

	start := time.Now()
	eintraege, err := handler.llm.ParseLebensmittelEingabe(request.Text)
	if err != nil {
		logging.Error("text-erfassung fehlgeschlagen",
			"text", logging.Kürzen(request.Text, 100), "fehler", err.Error())
		return fiber.NewError(fiber.StatusBadGateway, "Fehler bei der Text-Erfassung: "+err.Error())
	}
	if len(eintraege) == 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Im Text wurden keine Lebensmittel erkannt")
	}
	logging.Info("text-erfassung geparst",
		"text", logging.Kürzen(request.Text, 100),
		"erkannt", len(eintraege),
		"dauer_ms", time.Since(start).Milliseconds(),
	)

	// Bestehendes Inventar laden (für Merge und Gefrierfach-Erkennung)
	var inventar []models.Lebensmittel
	if err := handler.db.Find(&inventar).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Abrufen des Inventars: "+err.Error())
	}

	ergebnis := make([]models.Lebensmittel, 0, len(eintraege))
	for _, eintrag := range eintraege {
		var vorhandenes *models.Lebensmittel
		for i := range inventar {
			if strings.EqualFold(inventar[i].Name, eintrag.Name) {
				vorhandenes = &inventar[i]
				break
			}
		}

		if vorhandenes != nil {
			// Menge aufaddieren (mit Einheiten-Umrechnung)
			if gesamt, einheit, ok := addiereMengen(vorhandenes.Anzahl, vorhandenes.Einheit, eintrag.Anzahl, eintrag.Einheit); ok {
				vorhandenes.Anzahl = gesamt
				vorhandenes.Einheit = einheit
			}
			// Gefrierfach-Hinweis aus dem Text übernehmen
			if eintrag.Gefrierfach {
				vorhandenes.Gefrierfach = true
			}
			if err := handler.db.Save(vorhandenes).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Aktualisieren: "+err.Error())
			}
			ergebnis = append(ergebnis, *vorhandenes)
			continue
		}

		// Neuer Artikel: Kategorien per Regeln, Gefrierfach aus dem Text
		kategorie, mahlzeiten := models.KlassifiziereRegeln(eintrag.Name)
		neu := models.Lebensmittel{
			Name:                  eintrag.Name,
			Anzahl:                eintrag.Anzahl,
			Einheit:               eintrag.Einheit,
			LebensmittelKategorie: kategorie,
			Mahlzeiten:            mahlzeiten,
			Gefrierfach:           eintrag.Gefrierfach,
		}
		if err := handler.db.Create(&neu).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Erstellen: "+err.Error())
		}
		inventar = append(inventar, neu)
		ergebnis = append(ergebnis, neu)
	}

	names := make([]string, 0, len(ergebnis))
	for _, lm := range ergebnis {
		names = append(names, fmt.Sprintf("%s (%d %s)", lm.Name, lm.Anzahl, lm.Einheit))
	}
	logging.Info("lebensmittel per text erfasst",
		"anzahl", len(ergebnis),
		"artikel", strings.Join(names, ", "),
	)

	return ctx.JSON(convertLebensmittelSliceToListResponse(ergebnis))
}
// KategorisierenLebensmittel (POST /lebensmittel/kategorisieren)
// ordnet allen noch nicht kategorisierten Lebensmitteln eine Küchenrichtung
// und passende Mahlzeiten zu:
//  1. per Regel-Klassifikator (sofort, ohne API-Aufruf)
//  2. verbleibende per KI - in EINEM einzigen Batch-Aufruf für alle Artikel
//
// performant: unabhängig von der Inventargröße höchstens ein LLM-Aufruf.
func (handler *RezeptHandler) KategorisierenLebensmittel(ctx fiber.Ctx) error {
	var alle []models.Lebensmittel
	if err := handler.db.Find(&alle).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Abrufen des Inventars: "+err.Error())
	}

	perRegeln := 0
	perLLM := 0
	offen := 0

	// Pass 1: Regel-Klassifikator für alle ohne Zuordnung
	var rest []models.Lebensmittel
	for i := range alle {
		lm := &alle[i]
		if lm.LebensmittelKategorie != nil && len(lm.Mahlzeiten) > 0 {
			continue // bereits zugeordnet
		}
		kategorie, mahlzeiten := models.KlassifiziereRegeln(lm.Name)
		if kategorie != nil || len(mahlzeiten) > 0 {
			if kategorie != nil {
				lm.LebensmittelKategorie = kategorie
			}
			if len(mahlzeiten) > 0 {
				lm.Mahlzeiten = mahlzeiten
			}
			if err := handler.db.Save(lm).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Speichern: "+err.Error())
			}
			perRegeln++
		} else {
			rest = append(rest, *lm)
		}
	}

	// Pass 2: verbleibende per KI-Batch (ein Aufruf für alle)
	if len(rest) > 0 && handler.llm.Verfügbar() {
		eintraege := make([]llm.KlassifiziereAnfrageEintrag, len(rest))
		for i, lm := range rest {
			eintraege[i] = llm.KlassifiziereAnfrageEintrag{ID: lm.ID, Name: lm.Name}
		}

		zuordnungen, err := handler.llm.KlassifiziereLebensmittel(eintraege)
		if err != nil {
			logging.Error("ki-kategorisierung fehlgeschlagen", "artikel", len(eintraege), "fehler", err.Error())
			return fiber.NewError(fiber.StatusBadGateway, "Fehler bei der KI-Kategorisierung: "+err.Error())
		}

		for i := range rest {
			lm := &rest[i]
			z, ok := zuordnungen[lm.ID]
			if !ok {
				offen++
				continue
			}
			geändert := false
			if k := parseLebensmittelKategorie(z.LebensmittelKategorie); k != nil {
				lm.LebensmittelKategorie = k
				geändert = true
			}
			if m := parseMahlzeiten(z.Mahlzeiten); len(m) > 0 {
				lm.Mahlzeiten = m
				geändert = true
			}
			if geändert {
				if err := handler.db.Save(lm).Error; err != nil {
					return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Speichern: "+err.Error())
				}
				perLLM++
			} else {
				offen++
			}
		}
	} else if len(rest) > 0 {
		offen = len(rest)
	}

	logging.Info("kategorisierung abgeschlossen",
		"per_regeln", perRegeln, "per_llm", perLLM, "offen", offen)

	return ctx.JSON(schema.KategorisierenResponse{
		Success:   true,
		Message:   "Kategorisierung abgeschlossen",
		PerRegeln: perRegeln,
		PerLLM:    perLLM,
		Offen:     offen,
	})
}

// DeleteRezeptByID (DELETE /rezepte/:id) löscht ein Rezept aus der DB.
func (handler *RezeptHandler) DeleteRezeptByID(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	var rezept models.Rezept
	if err := handler.db.First(&rezept, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Rezept nicht gefunden")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Laden des Rezepts: "+err.Error())
	}

	if err := handler.db.Delete(&rezept).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Löschen: "+err.Error())
	}

	logging.Info("rezept gelöscht", "id", id, "name", rezept.Name)

	return ctx.JSON(schema.DeleteRezeptResponse{
		Success: true,
		Message: "Rezept erfolgreich gelöscht",
	})
}
