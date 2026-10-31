package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
	"gitlab.com/ReneGumzej/kitchenops/pkg/schema"
)

// setupEinkaufsApp erstellt eine App mit den Einkaufslisten-Routen.
func setupEinkaufsApp(db *gorm.DB) *fiber.App {
	// setupTestDB migriert Einkaufsartikel nicht - hier nachholen
	if err := db.AutoMigrate(&models.Einkaufsartikel{}); err != nil {
		panic("Fehler beim Migrieren: " + err.Error())
	}

	app := fiber.New()
	dbHandler := NewDatabaseHandler(db)

	app.Post("/einkaufsliste", dbHandler.CreateEinkaufsartikel)
	app.Put("/einkaufsliste/:id", dbHandler.UpdateEinkaufsartikel)

	return app
}

// abhaken markiert einen Artikel per PUT als erledigt und liefert die Antwort.
func abhaken(t *testing.T, app *fiber.App, id int, name string, anzahl int, einheit string) schema.EinkaufsartikelResponse {
	t.Helper()
	body, _ := json.Marshal(schema.UpdateEinkaufsartikelRequest{
		Name: name, Anzahl: uint16(anzahl), Einheit: einheit, Notiz: "", Erledigt: true,
	})
	req := httptest.NewRequest("PUT", fmt.Sprintf("/einkaufsliste/%d", id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response schema.EinkaufsartikelResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	return response
}

// TestAbhakenLegtNeuesLebensmittelAn: unbekannter Artikel wird beim Abhaken
// mit Regel-Kategorien im Inventar angelegt.
func TestAbhakenLegtNeuesLebensmittelAn(t *testing.T) {
	db := setupTestDB()
	app := setupEinkaufsApp(db)

	db.Create(&models.Einkaufsartikel{Name: "Schnittlauch", Anzahl: 1, Einheit: "Bund"})

	response := abhaken(t, app, 1, "Schnittlauch", 1, "Bund")
	assert.Contains(t, response.InventarHinweis, "Neu im Inventar: Schnittlauch")

	var lm models.Lebensmittel
	db.First(&lm, "name = ?", "Schnittlauch")
	assert.Equal(t, uint16(1), lm.Anzahl)
	assert.Equal(t, "Bund", lm.Einheit)
}

// TestAbhakenAddiertBestand: bekannter Artikel gleicher Einheit wird aufaddiert.
func TestAbhakenAddiertBestand(t *testing.T) {
	db := setupTestDB()
	app := setupEinkaufsApp(db)

	db.Create(&models.Lebensmittel{Name: "Mehl", Anzahl: 500, Einheit: "g"})
	db.Create(&models.Einkaufsartikel{Name: "Mehl", Anzahl: 200, Einheit: "g"})

	response := abhaken(t, app, 1, "Mehl", 200, "g")
	assert.Contains(t, response.InventarHinweis, "neuer Bestand 700 g")

	var lm models.Lebensmittel
	db.First(&lm, "name = ?", "Mehl")
	assert.Equal(t, uint16(700), lm.Anzahl)
	assert.Equal(t, 1, int(db.Find(&[]models.Lebensmittel{}).RowsAffected))
}

// TestAbhakenRechnetEinheitenUm: 2 kg Bestand + 500 g Einkauf = 2500 g.
func TestAbhakenRechnetEinheitenUm(t *testing.T) {
	db := setupTestDB()
	app := setupEinkaufsApp(db)

	db.Create(&models.Lebensmittel{Name: "Zucker", Anzahl: 2, Einheit: "kg"})
	db.Create(&models.Einkaufsartikel{Name: "Zucker", Anzahl: 500, Einheit: "g"})

	response := abhaken(t, app, 1, "Zucker", 500, "g")
	assert.Contains(t, response.InventarHinweis, "neuer Bestand 2500 g")

	var lm models.Lebensmittel
	db.First(&lm, "name = ?", "Zucker")
	assert.Equal(t, uint16(2500), lm.Anzahl)
	assert.Equal(t, "g", lm.Einheit)
}

// TestAbhakenInkompatibleEinheit: "Bund" zu "g" ist nicht umrechenbar -
// das Inventar bleibt unangetastet, der Hinweis erklärt es.
func TestAbhakenInkompatibleEinheit(t *testing.T) {
	db := setupTestDB()
	app := setupEinkaufsApp(db)

	db.Create(&models.Lebensmittel{Name: "Petersilie", Anzahl: 1, Einheit: "Bund"})
	db.Create(&models.Einkaufsartikel{Name: "Petersilie", Anzahl: 100, Einheit: "g"})

	response := abhaken(t, app, 1, "Petersilie", 100, "g")
	assert.Contains(t, response.InventarHinweis, "passt nicht")

	var lm models.Lebensmittel
	db.First(&lm, "name = ?", "Petersilie")
	assert.Equal(t, uint16(1), lm.Anzahl)
	assert.Equal(t, "Bund", lm.Einheit)
}

// TestAbhakenNurBeimUebergang: erneutes Abhaken eines bereits erledigten
// Artikels löst keine zweite Übernahme aus.
func TestAbhakenNurBeimUebergang(t *testing.T) {
	db := setupTestDB()
	app := setupEinkaufsApp(db)

	db.Create(&models.Lebensmittel{Name: "Eier", Anzahl: 6, Einheit: "Stück"})
	db.Create(&models.Einkaufsartikel{Name: "Eier", Anzahl: 6, Einheit: "Stück", Erledigt: true})

	response := abhaken(t, app, 1, "Eier", 6, "Stück")
	assert.Empty(t, response.InventarHinweis)

	var lm models.Lebensmittel
	db.First(&lm, "name = ?", "Eier")
	assert.Equal(t, uint16(6), lm.Anzahl)
}

// TestAbhakenLegtMitKategorienAn: neu angelegte Lebensmittel bekommen
// automatisch die Regel-Klassifizierung.
func TestAbhakenLegtMitKategorienAn(t *testing.T) {
	db := setupTestDB()
	app := setupEinkaufsApp(db)

	db.Create(&models.Einkaufsartikel{Name: "Rinderhackfleisch", Anzahl: 500, Einheit: "g"})

	response := abhaken(t, app, 1, "Rinderhackfleisch", 500, "g")
	assert.Contains(t, response.InventarHinweis, "Neu im Inventar")

	var lm models.Lebensmittel
	db.First(&lm, "name = ?", "Rinderhackfleisch")
	assert.NotNil(t, lm.LebensmittelKategorie)
	assert.Equal(t, models.LebensmittelKategorieRindfleisch, *lm.LebensmittelKategorie)
}
