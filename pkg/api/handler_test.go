package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
	"gitlab.com/ReneGumzej/kitchenops/pkg/schema"
)

// setupTestDB erstellt eine temporäre SQLite-Datenbank für Tests
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Fehler beim Erstellen der Test-Datenbank: " + err.Error())
	}

	// Auto-Migrate der Tabellen
	err = db.AutoMigrate(&models.Lebensmittel{}, &models.Sonstiges{}, &models.Rezept{})
	if err != nil {
		panic("Fehler beim Migrieren der Test-Datenbank: " + err.Error())
	}

	return db
}

// createTestApp erstellt eine Fiber-App mit dem DatabaseHandler für Tests
func createTestApp(db *gorm.DB) *fiber.App {
	app := fiber.New()
	dbHandler := NewDatabaseHandler(db)

	// Routen registrieren
	app.Get("/lebensmittel", dbHandler.GetAllLebensmittel)
	app.Get("/lebensmittel/:id", dbHandler.GetLebensmittelByID)
	app.Post("/lebensmittel", dbHandler.CreateNewLebensmittel)
	app.Delete("/lebensmittel/:id", dbHandler.DeleteLebensmittelByID)

	return app
}

// =============================================================================
// Tests für GetAllLebensmittel
// =============================================================================

func TestGetAllLebensmittel_Empty(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("GET", "/lebensmittel", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response schema.LebensmittelListResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Empty(t, response.Lebensmittel)
}

func TestGetAllLebensmittel_WithData(t *testing.T) {
	// Arrange
	db := setupTestDB()

	// Testdaten einfügen
	lebensmittel1 := models.Lebensmittel{Name: "Mehl", Anzahl: 500, Einheit: "g"}
	lebensmittel2 := models.Lebensmittel{Name: "Zucker", Anzahl: 250, Einheit: "g"}
	db.Create(&lebensmittel1)
	db.Create(&lebensmittel2)

	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("GET", "/lebensmittel", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response schema.LebensmittelListResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Lebensmittel, 2)

	// Prüfe, ob die Daten korrekt sind
	assert.Equal(t, "Mehl", response.Lebensmittel[0].Name)
	assert.Equal(t, uint16(500), response.Lebensmittel[0].Anzahl)
	assert.Equal(t, "g", response.Lebensmittel[0].Einheit)
}

// =============================================================================
// Tests für GetLebensmittelByID
// =============================================================================

func TestGetLebensmittelByID_Found(t *testing.T) {
	// Arrange
	db := setupTestDB()

	// Testdatum einfügen
	lebensmittel := models.Lebensmittel{Name: "Salz", Anzahl: 10, Einheit: "g"}
	db.Create(&lebensmittel)

	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("GET", "/lebensmittel/1", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response schema.LebensmittelResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Salz", response.Name)
	assert.Equal(t, uint16(10), response.Anzahl)
	assert.Equal(t, "g", response.Einheit)
}

func TestGetLebensmittelByID_NotFound(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("GET", "/lebensmittel/999", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetLebensmittelByID_InvalidID(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("GET", "/lebensmittel/invalid", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// =============================================================================
// Tests für CreateNewLebensmittel
// =============================================================================

func TestCreateNewLebensmittel_Success(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Request-Body vorbereiten
	newLebensmittel := schema.NewLebensmittelRequest{
		Name:    "Butter",
		Anzahl:  250,
		Einheit: "g",
	}
	body, _ := json.Marshal(newLebensmittel)

	// Act
	req := httptest.NewRequest("POST", "/lebensmittel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response schema.LebensmittelResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Butter", response.Name)
	assert.Equal(t, uint16(250), response.Anzahl)
	assert.Equal(t, "g", response.Einheit)
	assert.NotZero(t, response.ID) // ID sollte automatisch vergeben werden

	// Prüfen, ob das Lebensmittel in der DB ist
	var dbLebensmittel models.Lebensmittel
	db.First(&dbLebensmittel, response.ID)
	assert.Equal(t, "Butter", dbLebensmittel.Name)
}

func TestCreateNewLebensmittel_InvalidData(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Ungültiger Request-Body (fehlendes Feld)
	invalidBody := []byte(`{"name": "Test"}`)

	// Act
	req := httptest.NewRequest("POST", "/lebensmittel", bytes.NewReader(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreateNewLebensmittel_EmptyBody(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("POST", "/lebensmittel", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// =============================================================================
// Tests für DeleteLebensmittelByID
// =============================================================================

func TestDeleteLebensmittelByID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB()

	// Testdatum einfügen
	lebensmittel := models.Lebensmittel{Name: "Pfeffer", Anzahl: 5, Einheit: "g"}
	db.Create(&lebensmittel)

	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("DELETE", "/lebensmittel/1", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response schema.DeleteLebensmittelResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Lebensmittel erfolgreich gelöscht", response.Message)

	// Prüfen, ob das Lebensmittel wirklich gelöscht wurde
	var deletedLebensmittel models.Lebensmittel
	result := db.First(&deletedLebensmittel, 1)
	assert.Error(t, result.Error) // Sollte "record not found" Fehler geben
	assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
}

func TestDeleteLebensmittelByID_NotFound(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("DELETE", "/lebensmittel/999", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDeleteLebensmittelByID_InvalidID(t *testing.T) {
	// Arrange
	db := setupTestDB()
	app := createTestApp(db)

	// Act
	req := httptest.NewRequest("DELETE", "/lebensmittel/invalid", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
