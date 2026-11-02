package api

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gitlab.com/ReneGumzej/kitchenops/pkg/llm"
	"gitlab.com/ReneGumzej/kitchenops/pkg/logging"
	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
	"gitlab.com/ReneGumzej/kitchenops/pkg/schema"
	"gitlab.com/ReneGumzej/kitchenops/pkg/utils"
	"gorm.io/gorm"
)

// Handlers

type DatabaseHandler struct {
	db  *gorm.DB
	llm *llm.Client
}

func NewDatabaseHandler(db *gorm.DB, client *llm.Client) *DatabaseHandler {
	return &DatabaseHandler{
		db:  db,
		llm: client,
	}
}

// convertLebensmittelToResponse converts a models.Lebensmittel to a schema.LebensmittelResponse
func convertLebensmittelToResponse(lm models.Lebensmittel) schema.LebensmittelResponse {
	kategorie := ""
	if lm.LebensmittelKategorie != nil {
		kategorie = string(*lm.LebensmittelKategorie)
	}
	mahlzeiten := make([]string, len(lm.Mahlzeiten))
	for i, m := range lm.Mahlzeiten {
		mahlzeiten[i] = string(m)
	}
	return schema.LebensmittelResponse{
		ID:      lm.ID,
		Name:    lm.Name,
		Anzahl:  lm.Anzahl,
		Einheit: lm.Einheit,
		LebensmittelKategorie: kategorie,
		Mahlzeiten: mahlzeiten,
		Gefrierfach: lm.Gefrierfach,
	}
}

// gültigeKategorien für die Validierung von Request-Eingaben
var gültigeLebensmittelKategorien = map[string]bool{
	"Pasta": true, "Rindfleisch": true, "Hühnerfleisch": true, "Schweinefleisch": true,
	"Vegetarisch": true, "Vegan": true, "Salat": true,
}

var gültigeMahlzeiten = map[string]bool{
	"Frühstück": true, "Mittagessen": true, "Abendessen": true, "Snack": true,
}

// parseLebensmittelKategorie wandelt einen Request-String in eine validierte
// Kategorie um. Leer/ungültig ergibt nil (= keine Zuordnung).
func parseLebensmittelKategorie(wert string) *models.LebensmittelKategorie {
	if !gültigeLebensmittelKategorien[wert] {
		return nil
	}
	k := models.LebensmittelKategorie(wert)
	return &k
}

// parseMahlzeiten filtert eine Request-Liste auf gültige Mahlzeiten.
func parseMahlzeiten(werte []string) []models.EssensKategorie {
	result := make([]models.EssensKategorie, 0, len(werte))
	for _, w := range werte {
		if gültigeMahlzeiten[w] {
			result = append(result, models.EssensKategorie(w))
		}
	}
	return result
}

// convertLebensmittelSliceToListResponse converts a slice of models.Lebensmittel to a schema.LebensmittelListResponse
func convertLebensmittelSliceToListResponse(lebensmittel []models.Lebensmittel) schema.LebensmittelListResponse {
	responses := make([]schema.LebensmittelResponse, len(lebensmittel))
	for i, lm := range lebensmittel {
		responses[i] = convertLebensmittelToResponse(lm)
	}
	return schema.LebensmittelListResponse{
		Lebensmittel: responses,
	}
}


// Functions

func InitGreet(ctx fiber.Ctx) error {
	return ctx.SendString(utils.BANNER)
}
func (handler *DatabaseHandler) GetAllLebensmittel(ctx fiber.Ctx) error {
	var lebensmittel []models.Lebensmittel

	if err := handler.db.Find(&lebensmittel).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Abrufen der Lebensmittel: "+err.Error())
	}

	responses := convertLebensmittelSliceToListResponse(lebensmittel)

	return ctx.JSON(responses)
}


func (handler *DatabaseHandler) GetLebensmittelByID(ctx fiber.Ctx) error {
	var lebensmittel models.Lebensmittel

	id := ctx.Params("id")
	if err := handler.db.First(&lebensmittel, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Lebensmittel nicht gefunden")
	}

	return ctx.JSON(convertLebensmittelToResponse(lebensmittel))
}

func (handler *DatabaseHandler) CreateNewLebensmittel(ctx fiber.Ctx) error {
	newLebensmittel := &schema.NewLebensmittelRequest{}
	
	if err := ctx.Bind().Body(newLebensmittel); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}

	// Validierung: Alle Pflichtfelder müssen gesetzt sein
	if newLebensmittel.Name == "" || newLebensmittel.Einheit == "" || newLebensmittel.Anzahl == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Name, Anzahl und Einheit sind Pflichtfelder")
	}

	// Kategorien: explizit aus dem Request, sonst Regel-Klassifikator
	kategorie := parseLebensmittelKategorie(newLebensmittel.LebensmittelKategorie)
	mahlzeiten := parseMahlzeiten(newLebensmittel.Mahlzeiten)
	if kategorie == nil || len(mahlzeiten) == 0 {
		regelKategorie, regelMahlzeiten := models.KlassifiziereRegeln(newLebensmittel.Name)
		if kategorie == nil {
			kategorie = regelKategorie
		}
		if len(mahlzeiten) == 0 {
			mahlzeiten = regelMahlzeiten
		}
	}

	lebensmittelTable := models.Lebensmittel{
		Name: newLebensmittel.Name,
		Anzahl: newLebensmittel.Anzahl,
		Einheit: newLebensmittel.Einheit,
		LebensmittelKategorie: kategorie,
		Mahlzeiten: mahlzeiten,
		Gefrierfach: newLebensmittel.Gefrierfach,
	}
	if err := handler.db.Create(&lebensmittelTable).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Erstellen: "+err.Error())
	}

	logging.Info("lebensmittel angelegt",
		"id", lebensmittelTable.ID,
		"name", lebensmittelTable.Name,
		"anzahl", lebensmittelTable.Anzahl,
		"einheit", lebensmittelTable.Einheit,
		"gefrierfach", lebensmittelTable.Gefrierfach,
	)

	return ctx.JSON(convertLebensmittelToResponse(lebensmittelTable))
}

func (handler *DatabaseHandler) DeleteLebensmittelByID(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	var lebensmittel models.Lebensmittel
	if err := handler.db.First(&lebensmittel, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Lebensmittel nicht gefunden")
	}

	if err := handler.db.Delete(&lebensmittel).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Löschen: "+err.Error())
	}

	logging.Info("lebensmittel gelöscht", "id", lebensmittel.ID, "name", lebensmittel.Name)

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Lebensmittel erfolgreich gelöscht",
	})
}

func (handler *DatabaseHandler) UpdateLebensmittelByID(ctx fiber.Ctx) error {
	var lebensmittelTable models.Lebensmittel
	
	lebenmittelRequest := &schema.UpdateLebensmittelRequest{}

	if err := ctx.Bind().Body(lebenmittelRequest); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	id := ctx.Params("id")

	if err := handler.db.Find(&lebensmittelTable, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	lebensmittelTable.Name = lebenmittelRequest.Name
	lebensmittelTable.Anzahl = lebenmittelRequest.Anzahl
	lebensmittelTable.Einheit = lebenmittelRequest.Einheit

	// Kategorien: nur überschreiben, wenn im Request angegeben;
	// andernfalls bestehende Zuordnung behalten (oder per Regeln ermitteln)
	if kategorie := parseLebensmittelKategorie(lebenmittelRequest.LebensmittelKategorie); kategorie != nil {
		lebensmittelTable.LebensmittelKategorie = kategorie
	} else if lebensmittelTable.LebensmittelKategorie == nil {
		if regelKategorie, _ := models.KlassifiziereRegeln(lebenmittelRequest.Name); regelKategorie != nil {
			lebensmittelTable.LebensmittelKategorie = regelKategorie
		}
	}
	if mahlzeiten := parseMahlzeiten(lebenmittelRequest.Mahlzeiten); len(mahlzeiten) > 0 {
		lebensmittelTable.Mahlzeiten = mahlzeiten
	} else if len(lebensmittelTable.Mahlzeiten) == 0 {
		if _, regelMahlzeiten := models.KlassifiziereRegeln(lebenmittelRequest.Name); len(regelMahlzeiten) > 0 {
			lebensmittelTable.Mahlzeiten = regelMahlzeiten
		}
	}
	// Gefrierfach nur überschreiben, wenn das Feld im Request ankommt
	if lebenmittelRequest.Gefrierfach != nil {
		lebensmittelTable.Gefrierfach = *lebenmittelRequest.Gefrierfach
	}

	if err := handler.db.Save(&lebensmittelTable).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	logging.Info("lebensmittel aktualisiert",
		"id", lebensmittelTable.ID,
		"name", lebensmittelTable.Name,
		"anzahl", lebensmittelTable.Anzahl,
		"einheit", lebensmittelTable.Einheit,
	)

	return ctx.JSON(convertLebensmittelToResponse(lebensmittelTable))
}

// =============================================================================
// Einkaufsliste
// =============================================================================

// convertEinkaufsartikelToResponse converts a models.Einkaufsartikel to a schema.EinkaufsartikelResponse
func convertEinkaufsartikelToResponse(a models.Einkaufsartikel) schema.EinkaufsartikelResponse {
	return schema.EinkaufsartikelResponse{
		ID:          a.ID,
		Name:        a.Name,
		Anzahl:      a.Anzahl,
		Einheit:     a.Einheit,
		Notiz:       a.Notiz,
		Erledigt:    a.Erledigt,
		Gefrierfach: a.Gefrierfach,
	}
}

// convertEinkaufsartikelSliceToListResponse converts a slice of models.Einkaufsartikel to a schema.EinkaufsartikelListResponse
func convertEinkaufsartikelSliceToListResponse(artikel []models.Einkaufsartikel) schema.EinkaufsartikelListResponse {
	responses := make([]schema.EinkaufsartikelResponse, len(artikel))
	for i, a := range artikel {
		responses[i] = convertEinkaufsartikelToResponse(a)
	}
	return schema.EinkaufsartikelListResponse{
		Artikel: responses,
	}
}

// GetAllEinkaufsartikel returns all entries of the Einkaufsliste.
// Open (not erledigt) entries first, then sorted by name.
func (handler *DatabaseHandler) GetAllEinkaufsartikel(ctx fiber.Ctx) error {
	var artikel []models.Einkaufsartikel

	if err := handler.db.
		Order("erledigt asc").
		Order("name asc").
		Find(&artikel).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Abrufen der Einkaufsliste: "+err.Error())
	}

	return ctx.JSON(convertEinkaufsartikelSliceToListResponse(artikel))
}

// CreateEinkaufsartikel adds a new entry to the Einkaufsliste.
// If an open entry with the same name already exists, its amount, unit and note are updated instead.
func (handler *DatabaseHandler) CreateEinkaufsartikel(ctx fiber.Ctx) error {
	newArtikel := &schema.NewEinkaufsartikelRequest{}

	if err := ctx.Bind().Body(newArtikel); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}

	// Validierung: Name und Einheit sind Pflichtfelder
	if newArtikel.Name == "" || newArtikel.Einheit == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name und Einheit sind Pflichtfelder")
	}

	// Anzahl: falls nicht angegeben, auf 1 setzen
	anzahl := newArtikel.Anzahl
	if anzahl == 0 {
		anzahl = 1
	}

	// Dubletten vermeiden: offenen Eintrag mit gleichem Namen aktualisieren
	var existing models.Einkaufsartikel
	err := handler.db.
		Where("name = ? AND erledigt = ?", newArtikel.Name, false).
		First(&existing).Error
	if err == nil {
		existing.Anzahl = anzahl
		existing.Einheit = newArtikel.Einheit
		if newArtikel.Notiz != "" {
			existing.Notiz = newArtikel.Notiz
		}
		if err := handler.db.Save(&existing).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Aktualisieren des Eintrags: "+err.Error())
		}
		logging.Info("einkaufsartikel aktualisiert (dublette)",
			"id", existing.ID, "name", existing.Name, "anzahl", existing.Anzahl, "einheit", existing.Einheit)
		return ctx.JSON(convertEinkaufsartikelToResponse(existing))
	}

	artikel := models.Einkaufsartikel{
		Name:        newArtikel.Name,
		Anzahl:      anzahl,
		Einheit:     newArtikel.Einheit,
		Notiz:       newArtikel.Notiz,
		Erledigt:    false,
		Gefrierfach: newArtikel.Gefrierfach,
	}
	if err := handler.db.Create(&artikel).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Erstellen: "+err.Error())
	}

	logging.Info("einkaufsartikel angelegt",
		"id", artikel.ID, "name", artikel.Name,
		"anzahl", artikel.Anzahl, "einheit", artikel.Einheit, "gefrierfach", artikel.Gefrierfach)

	return ctx.JSON(convertEinkaufsartikelToResponse(artikel))
}

// UpdateEinkaufsartikel updates an entry of the Einkaufsliste (also used to toggle "erledigt").
func (handler *DatabaseHandler) UpdateEinkaufsartikel(ctx fiber.Ctx) error {
	updateRequest := &schema.UpdateEinkaufsartikelRequest{}

	if err := ctx.Bind().Body(updateRequest); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}

	if updateRequest.Name == "" || updateRequest.Einheit == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name und Einheit sind Pflichtfelder")
	}

	id := ctx.Params("id")

	var artikel models.Einkaufsartikel
	if err := handler.db.First(&artikel, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Eintrag nicht gefunden")
	}

	// Abhaken abfragen, um die Automatik nur beim Übergang auszulösen
	warErledigt := artikel.Erledigt

	artikel.Name = updateRequest.Name
	artikel.Anzahl = updateRequest.Anzahl
	artikel.Einheit = updateRequest.Einheit
	artikel.Notiz = updateRequest.Notiz
	artikel.Erledigt = updateRequest.Erledigt
	// Gefrierfach nur überschreiben, wenn das Feld im Request ankommt
	if updateRequest.Gefrierfach != nil {
		artikel.Gefrierfach = *updateRequest.Gefrierfach
	}

	if err := handler.db.Save(&artikel).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Aktualisieren: "+err.Error())
	}

	// Automatik: abgehakter Artikel wandert ins Inventar
	inventarHinweis := ""
	if !warErledigt && artikel.Erledigt {
		inventarHinweis = handler.übernehmeInInventar(artikel)
	}

	if inventarHinweis != "" {
		logging.Info("einkauf abgehakt - inventar übernommen",
			"name", artikel.Name, "hinweis", inventarHinweis)
	} else {
		logging.Info("einkaufsartikel aktualisiert",
			"id", artikel.ID, "name", artikel.Name, "erledigt", artikel.Erledigt)
	}

	response := convertEinkaufsartikelToResponse(artikel)
	response.InventarHinweis = inventarHinweis
	return ctx.JSON(response)
}

// übernehmeInInventar führt einen abgehakten Einkaufsartikel ins Inventar über:
// ein vorhandenes Lebensmittel gleichen Namens wird aufaddiert (mit
// Einheiten-Umrechnung kg/g und l/ml), ein unbekanntes wird neu angelegt.
// Rückgabe ist ein Hinweistext für die UI.
func (handler *DatabaseHandler) übernehmeInInventar(artikel models.Einkaufsartikel) string {
	var inventar []models.Lebensmittel
	if err := handler.db.Find(&inventar).Error; err != nil {
		return ""
	}

	for i := range inventar {
		lm := &inventar[i]
		if !strings.EqualFold(lm.Name, artikel.Name) {
			continue
		}

		gesamt, einheit, ok := addiereMengen(lm.Anzahl, lm.Einheit, artikel.Anzahl, artikel.Einheit)
		if !ok {
			return fmt.Sprintf("Nicht ins Inventar übernommen: Einheit %s passt nicht zu bestehendem %s (%s) - bitte manuell anpassen",
				artikel.Einheit, lm.Name, lm.Einheit)
		}
		lm.Anzahl = gesamt
		lm.Einheit = einheit
		if err := handler.db.Save(lm).Error; err != nil {
			return ""
		}
		return fmt.Sprintf("Ins Inventar übernommen: %s - neuer Bestand %d %s", lm.Name, lm.Anzahl, lm.Einheit)
	}

	// Kein passendes Lebensmittel vorhanden -> neu anlegen
	// (Gefrierfach-Flag des Artikels wird übernommen)
	kategorie, mahlzeiten := models.KlassifiziereRegeln(artikel.Name)
	neu := models.Lebensmittel{
		Name:                  artikel.Name,
		Anzahl:                artikel.Anzahl,
		Einheit:               artikel.Einheit,
		LebensmittelKategorie: kategorie,
		Mahlzeiten:            mahlzeiten,
		Gefrierfach:            artikel.Gefrierfach,
	}
	if err := handler.db.Create(&neu).Error; err != nil {
		return ""
	}
	return fmt.Sprintf("Neu im Inventar: %s (%d %s)", neu.Name, neu.Anzahl, neu.Einheit)
}

// einheitsFamilien ordnet Einheiten einem Basismultiplikator zu
// (Faktor in der kleinsten Einheit der Familie).
var einheitsFamilien = []map[string]uint64{
	{"g": 1, "kg": 1000},
	{"ml": 1, "l": 1000},
}

// addiereMengen addiert zwei Mengenangaben. Passt die Einheit nicht und ist
// sie nicht umrechenbar, liefert ok=false. Bei Umrechnung wird das Ergebnis
// auf die feinere der beiden Einheiten normalisiert (z.B. 2 kg + 500 g = 2500 g).
func addiereMengen(basisAnzahl uint16, basisEinheit string, addAnzahl uint16, addEinheit string) (uint16, string, bool) {
	b := strings.ToLower(strings.TrimSpace(basisEinheit))
	a := strings.ToLower(strings.TrimSpace(addEinheit))

	if b == a {
		// Original-Schreibweise der Bestands-Einheit beibehalten
		return addiereSicher(basisAnzahl, addAnzahl), strings.TrimSpace(basisEinheit), true
	}

	for _, familie := range einheitsFamilien {
		fb, okB := familie[b]
		fa, okA := familie[a]
		if !okB || !okA {
			continue
		}
		// auf die feinere Einheit normalisieren (Original-Schreibweise)
		einheit := strings.TrimSpace(basisEinheit)
		if fa < fb {
			einheit = strings.TrimSpace(addEinheit)
		}
		gesamt := uint64(basisAnzahl)*fb + uint64(addAnzahl)*fa
		if gesamt > math.MaxUint16 {
			gesamt = math.MaxUint16
		}
		return uint16(gesamt), einheit, true
	}

	return 0, "", false
}

// addiereSicher addiert zwei uint16-Werte mit Überlaufschutz.
func addiereSicher(a, b uint16) uint16 {
	summe := uint64(a) + uint64(b)
	if summe > math.MaxUint16 {
		return math.MaxUint16
	}
	return uint16(summe)
}

// DeleteEinkaufsartikelByID removes a single entry from the Einkaufsliste.
func (handler *DatabaseHandler) DeleteEinkaufsartikelByID(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	var artikel models.Einkaufsartikel
	if err := handler.db.First(&artikel, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Eintrag nicht gefunden")
	}

	if err := handler.db.Delete(&artikel).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Löschen: "+err.Error())
	}

	logging.Info("einkaufsartikel gelöscht", "id", artikel.ID, "name", artikel.Name)

	return ctx.JSON(schema.DeleteEinkaufsartikelResponse{
		Success: true,
		Message: "Eintrag erfolgreich gelöscht",
	})
}

// ClearEinkaufsliste removes all entries from the Einkaufsliste.
func (handler *DatabaseHandler) ClearEinkaufsliste(ctx fiber.Ctx) error {
	result := handler.db.Where("1 = 1").Delete(&models.Einkaufsartikel{})
	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Leeren der Einkaufsliste: "+result.Error.Error())
	}

	logging.Info("einkaufsliste geleert", "gelöscht", result.RowsAffected)

	return ctx.JSON(schema.DeleteEinkaufsartikelResponse{
		Success: true,
		Message: fmt.Sprintf("Einkaufsliste geleert (%d Einträge gelöscht)", result.RowsAffected),
	})
}

// ErfasseEinkaufsartikelPerText (POST /einkaufsliste/erfassen) liest einen
// freien Text (z.B. "Füge mir Milch, Butter und zwei mal Sahne hinzu und packe die
// Eier ins Gefrierfach"), lässt ihn vom LLM strukturieren und trägt die Artikel
// in die Einkaufsliste ein. Verglichen mit ErfasseLebensmittelPerText:
// - Es werden Einkaufsartikel statt Lebensmittel angelegt
// - Keine automatische Gefrierfach-Übertragung ins Inventar
// - Gefrierfach-Flag wird aus dem Text übernommen
func (handler *DatabaseHandler) ErfasseEinkaufsartikelPerText(ctx fiber.Ctx) error {
	request := &schema.EinkaufzettelEingabeRequest{}

	if err := ctx.Bind().Body(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfragedaten: "+err.Error())
	}
	if request.Text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Bitte geben Sie einen Text ein")
	}
	if handler.llm == nil || !handler.llm.Verfügbar() {
		return fiber.NewError(fiber.StatusServiceUnavailable,
			"Text-Erfassung nicht konfiguriert: bitte einen API-Key für Ihren LLM-Anbieter als Umgebungsvariable setzen (z.B. MISTRAL_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY)")
	}

	start := time.Now()
	eintraege, err := handler.llm.ParseLebensmittelEingabe(request.Text)
	if err != nil {
		logging.Error("text-erfassung für einkaufsliste fehlgeschlagen",
			"text", logging.Kürzen(request.Text, 100), "fehler", err.Error())
		return fiber.NewError(fiber.StatusBadGateway, "Fehler bei der Text-Erfassung: "+err.Error())
	}
	if len(eintraege) == 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Im Text wurden keine Artikel erkannt")
	}
	logging.Info("text-erfassung für einkaufsliste geparst",
		"text", logging.Kürzen(request.Text, 100),
		"erkannt", len(eintraege),
		"dauer_ms", time.Since(start).Milliseconds(),
	)

	// Artikel in die Einkaufsliste eintragen
	ergebnis := make([]models.Einkaufsartikel, 0, len(eintraege))
	for _, eintrag := range eintraege {
		// Anzahl: 1 wenn nicht angegeben
		anzahl := eintrag.Anzahl
		if anzahl == 0 {
			anzahl = 1
		}

		// Dubletten vermeiden: offiziellen Eintrag mit gleichem Namen aktualisieren
		var existing models.Einkaufsartikel
		err := handler.db.
			Where("name = ? AND erledigt = ?", eintrag.Name, false).
			First(&existing).Error

		if err == nil {
			// Bestehender offener Artikel: Menge aufaddieren
			existing.Anzahl += anzahl
			if err := handler.db.Save(&existing).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Aktualisieren: "+err.Error())
			}
			ergebnis = append(ergebnis, existing)
			continue
		}

		// Neuer Artikel
		artikel := models.Einkaufsartikel{
			Name:        eintrag.Name,
			Anzahl:      anzahl,
			Einheit:     eintrag.Einheit,
			Notiz:       "",
			Erledigt:    false,
			Gefrierfach: eintrag.Gefrierfach,
		}
		if err := handler.db.Create(&artikel).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Erstellen: "+err.Error())
		}
		ergebnis = append(ergebnis, artikel)
	}

	names := make([]string, 0, len(ergebnis))
	for _, a := range ergebnis {
		names = append(names, fmt.Sprintf("%s (%d %s)", a.Name, a.Anzahl, a.Einheit))
	}
	logging.Info("einkaufsartikel per text erfasst",
		"anzahl", len(ergebnis),
		"artikel", strings.Join(names, ", "),
	)

	return ctx.JSON(convertEinkaufsartikelSliceToListResponse(ergebnis))
}