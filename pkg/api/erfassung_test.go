package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.com/ReneGumzej/kitchenops/pkg/llm"
	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
	"gitlab.com/ReneGumzej/kitchenops/pkg/schema"
)

// mockErfassungsServer liefert eine canned LLM-Antwort für die Text-Erfassung.
// antwort ist das Objekt, das der LLM-Inhalt sein soll (z.B. {"lebensmittel": [...]}).
func mockErfassungsServer(t *testing.T, antwort map[string]any) *httptest.Server {
	t.Helper()

	// Antwortobjekt serialisieren und als JSON-String in den content einbetten
	inner, err := json.Marshal(antwort)
	assert.NoError(t, err)
	content, err := json.Marshal(string(inner))
	assert.NoError(t, err)
	completion := `{"choices":[{"message":{"content":` + string(content) + `}}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(completion))
	}))
	t.Setenv("MISTRAL_API_KEY", "test-key")
	t.Setenv("MISTRAL_BASE_URL", server.URL)
	return server
}

// setupErfassungsApp erstellt eine App mit der Text-Erfassungs-Route.
func setupErfassungsApp(db *gorm.DB, client *llm.Client) *fiber.App {
	app := fiber.New()
	handler := NewRezeptHandler(db, client)
	app.Post("/lebensmittel/erfassen", handler.ErfasseLebensmittelPerText)
	return app
}

// TestErfassungLegtNeueLebensmittelAn: Freitext wird in strukturierte
// Einträge zerlegt und in der DB angelegt (mit Regel-Kategorien).
func TestErfassungLegtNeueLebensmittelAn(t *testing.T) {
	db := setupTestDB()
	server := mockErfassungsServer(t, map[string]any{"lebensmittel": []map[string]any{
		{"name": "Brot", "anzahl": 1, "einheit": "Stück", "gefrierfach": false},
		{"name": "Sahne", "anzahl": 2, "einheit": "Becher", "gefrierfach": false},
		{"name": "Hähnchen", "anzahl": 1, "einheit": "Stück", "gefrierfach": true},
	}})
	defer server.Close()
	app := setupErfassungsApp(db, llm.NewClient())

	body, _ := json.Marshal(schema.LebensmittelEingabeRequest{
		Text: "Füge mir Brot, zwei mal Sahne und ein Hähnchen ins Gefrierfach hinzu",
	})
	req := httptest.NewRequest("POST", "/lebensmittel/erfassen", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result schema.LebensmittelListResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Len(t, result.Lebensmittel, 3)

	var brot, sahne, haehnchen models.Lebensmittel
	db.First(&brot, "name = ?", "Brot")
	db.First(&sahne, "name = ?", "Sahne")
	db.First(&haehnchen, "name = ?", "Hähnchen")

	assert.Equal(t, uint16(1), brot.Anzahl)
	assert.Equal(t, "Stück", brot.Einheit)
	assert.False(t, brot.Gefrierfach)

	assert.Equal(t, uint16(2), sahne.Anzahl)
	assert.Equal(t, "Becher", sahne.Einheit)

	assert.Equal(t, uint16(1), haehnchen.Anzahl)
	assert.True(t, haehnchen.Gefrierfach)
	// Regel-Klassifizierung: Hähnchen -> Hühnerfleisch
	assert.NotNil(t, haehnchen.LebensmittelKategorie)
	assert.Equal(t, models.LebensmittelKategorieHuehnerfleisch, *haehnchen.LebensmittelKategorie)
}

// TestErfassungAddiertBestand: existierende Lebensmittel werden aufaddiert
// statt doppelt angelegt.
func TestErfassungAddiertBestand(t *testing.T) {
	db := setupTestDB()
	db.Create(&models.Lebensmittel{Name: "Mehl", Anzahl: 500, Einheit: "g"})

	server := mockErfassungsServer(t, map[string]any{"lebensmittel": []map[string]any{
		{"name": "Mehl", "anzahl": 500, "einheit": "g", "gefrierfach": false},
	}})
	defer server.Close()
	app := setupErfassungsApp(db, llm.NewClient())

	body, _ := json.Marshal(schema.LebensmittelEingabeRequest{Text: "Füge mir 500 g Mehl hinzu"})
	req := httptest.NewRequest("POST", "/lebensmittel/erfassen", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var mehl models.Lebensmittel
	db.First(&mehl, "name = ?", "Mehl")
	assert.Equal(t, uint16(1000), mehl.Anzahl)
	var count int64
	db.Model(&models.Lebensmittel{}).Where("name = ?", "Mehl").Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestErfassungSetztGefrierfachAufBestand: Gefrierfach-Hinweis im Text
// setzt das Flag auch bei bereits vorhandenem Lebensmittel.
func TestErfassungSetztGefrierfachAufBestand(t *testing.T) {
	db := setupTestDB()
	db.Create(&models.Lebensmittel{Name: "Hähnchenbrust", Anzahl: 2, Einheit: "Stück"})

	server := mockErfassungsServer(t, map[string]any{"lebensmittel": []map[string]any{
		{"name": "Hähnchenbrust", "anzahl": 1, "einheit": "Stück", "gefrierfach": true},
	}})
	defer server.Close()
	app := setupErfassungsApp(db, llm.NewClient())

	body, _ := json.Marshal(schema.LebensmittelEingabeRequest{
		Text: "Füge mir Hähnchenbrust hinzu und packe sie ins Gefrierfach",
	})
	req := httptest.NewRequest("POST", "/lebensmittel/erfassen", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var haehnchen models.Lebensmittel
	db.First(&haehnchen, "name = ?", "Hähnchenbrust")
	assert.Equal(t, uint16(3), haehnchen.Anzahl)
	assert.True(t, haehnchen.Gefrierfach)
}

// TestErfassungOhneText: leerer Text wird mit 400 abgewiesen.
func TestErfassungOhneText(t *testing.T) {
	db := setupTestDB()
	server := mockErfassungsServer(t, map[string]any{"lebensmittel": []map[string]any{}})
	defer server.Close()
	app := setupErfassungsApp(db, llm.NewClient())

	body, _ := json.Marshal(schema.LebensmittelEingabeRequest{Text: ""})
	req := httptest.NewRequest("POST", "/lebensmittel/erfassen", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
