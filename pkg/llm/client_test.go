package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zendev-sh/goai"

	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
)

// mockServer liefert eine canned chat/completions-Antwort (OpenAI-kompatibles
// Format, das auch der Mistral-Provider von GoAI verwendet).
func mockServer(t *testing.T, antwort map[string]any) *httptest.Server {
	t.Helper()

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

// ============================================================================
// Prompt-Aufbau
// ============================================================================

// TestBuildUserPrompt prüft den strukturierten Aufbau der Nutzer-Anfrage.
func TestBuildUserPrompt(t *testing.T) {
	inventar := []models.Lebensmittel{
		{Name: "Mehl", Anzahl: 500, Einheit: "g"},
		{Name: "Eier", Anzahl: 6, Einheit: "Stück"},
	}

	anfrage := Anfrage{
		Prompt:                "ein vegetarisches Gericht",
		EssensKategorie:       "Frühstück",
		LebensmittelKategorie: "Vegetarisch",
		Anzahl:                2,
	}

	prompt := buildUserPrompt(inventar, anfrage)

	assert.Contains(t, prompt, "Essenskategorie (verbindlich): Frühstück")
	assert.Contains(t, prompt, "Lebensmittelkategorie (verbindlich): Vegetarisch")
	assert.Contains(t, prompt, "Anzahl Rezepte: 2")
	assert.Contains(t, prompt, "ein vegetarisches Gericht")
	assert.Contains(t, prompt, "- Mehl: 500 g")
	assert.Contains(t, prompt, "- Eier: 6 Stück")
}

// TestLeeresInventar prüft den Fallback bei leerem Inventar.
func TestLeeresInventar(t *testing.T) {
	prompt := buildUserPrompt(nil, Anfrage{Anzahl: 1})
	assert.Contains(t, prompt, "(leer")
}

// ============================================================================
// Schema-Generierung aus Structs (GoAI)
// ============================================================================

// TestSchemaAusStructs stellt sicher, dass die jsonschema-Tags der Typen
// ein vollständiges Schema mit Enums und Pflichtfeldern erzeugen.
func TestSchemaAusStructs(t *testing.T) {
	schema := goai.SchemaFrom[RezeptAntwort]()
	text := string(schema)

	assert.Contains(t, text, `"rezepte"`)
	assert.Contains(t, text, `"zubereitungszeit_minuten"`)
	assert.Contains(t, text, "Frühstück")
	assert.Contains(t, text, "Hühnerfleisch")
	assert.Contains(t, text, `"enum"`)
}

// ============================================================================
// Client-Fabrik (Provider-Erkennung, Standardmodelle)
// ============================================================================

// TestAnbieterErkennung: ohne LLM_PROVIDER entscheidet der vorhandene
// API-Key; Mistral hat als bisheriger Standard Vorrang.
func TestAnbieterErkennung(t *testing.T) {
	t.Run("Mistral-Key -> Mistral", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "k")
		t.Setenv("OPENAI_API_KEY", "")
		assert.Equal(t, ProviderMistral, NewClient().Anbieter())
	})

	t.Run("nur OpenAI-Key -> OpenAI", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "")
		t.Setenv("OPENAI_API_KEY", "sk-test")
		assert.Equal(t, ProviderOpenAI, NewClient().Anbieter())
	})

	t.Run("expliziter LLM_PROVIDER gewinnt", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "k")
		t.Setenv("LLM_PROVIDER", "anthropic")
		t.Setenv("ANTHROPIC_API_KEY", "k2")
		assert.Equal(t, ProviderAnthropic, NewClient().Anbieter())
	})
}

// TestModellKonfiguration: LLM_MODEL überschreibt, MISTRAL_MODEL bleibt
// als Abwärtskompatibilität wirksam, sonst gilt das Anbieter-Standardmodell.
func TestModellKonfiguration(t *testing.T) {
	t.Run("Standardmodell", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "k")
		t.Setenv("LLM_MODEL", "")
		t.Setenv("MISTRAL_MODEL", "")
		t.Setenv("LLM_PROVIDER", "")
		assert.Equal(t, "mistral-small-2603", NewClient().Model())
	})

	t.Run("MISTRAL_MODEL (Abwärtskompatibilität)", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "k")
		t.Setenv("LLM_MODEL", "")
		t.Setenv("MISTRAL_MODEL", "mistral-large-latest")
		t.Setenv("LLM_PROVIDER", "")
		assert.Equal(t, "mistral-large-latest", NewClient().Model())
	})

	t.Run("LLM_MODEL überschreibt", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "k")
		t.Setenv("LLM_MODEL", "open-mistral-nemo")
		t.Setenv("MISTRAL_MODEL", "mistral-large-latest")
		t.Setenv("LLM_PROVIDER", "")
		assert.Equal(t, "open-mistral-nemo", NewClient().Model())
	})
}

// TestOhneKeyVerfuegbarFalse: ohne jeden API-Key ist der Client nicht verfügbar.
func TestOhneKeyVerfuegbarFalse(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GOOGLE_GENERATIVE_AI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "")

	client := NewClient()
	assert.False(t, client.Verfügbar())

	_, err := client.GeneriereRezepte(nil, Anfrage{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kein API-Key")
}

// ============================================================================
// LLM-Funktionen gegen einen simulierten API-Server
// ============================================================================

// TestGeneriereRezepte prüft die komplette Client-Logik gegen einen
// simulierten Mistral-API-Server (structured output wird geparst).
func TestGeneriereRezepte(t *testing.T) {
	server := mockServer(t, map[string]any{"rezepte": []map[string]any{
		{
			"name":                   "Pfannkuchen",
			"beschreibung":           "Fluffig und schnell",
			"essens_kategorie":       "Frühstück",
			"lebensmittel_kategorie": "Vegetarisch",
			"zutaten": []map[string]any{
				{"name": "Mehl", "menge": "200", "einheit": "g", "vorhanden": true},
				{"name": "Milch", "menge": "300", "einheit": "ml", "vorhanden": false},
			},
			"schritte":                []string{"Teig anrühren", "Braten"},
			"zubereitungszeit_minuten": 15,
			"schwierigkeit":           "Einfach",
			"portionen":               2,
		},
	}})
	defer server.Close()

	client := NewClient()
	assert.True(t, client.Verfügbar())
	assert.Equal(t, "mistral-small-2603", client.Model())

	rezepte, err := client.GeneriereRezepte(
		[]models.Lebensmittel{{Name: "Mehl", Anzahl: 500, Einheit: "g"}},
		Anfrage{EssensKategorie: "Frühstück", Anzahl: 1},
	)

	assert.NoError(t, err)
	assert.Len(t, rezepte, 1)
	assert.Equal(t, "Pfannkuchen", rezepte[0].Name)
	assert.Equal(t, "Frühstück", rezepte[0].EssensKategorie)
	assert.Len(t, rezepte[0].Zutaten, 2)
	assert.True(t, rezepte[0].Zutaten[0].Vorhanden)
	assert.False(t, rezepte[0].Zutaten[1].Vorhanden)
	assert.Equal(t, uint16(15), rezepte[0].ZubereitungszeitMin)
	assert.Equal(t, []string{"Teig anrühren", "Braten"}, rezepte[0].Schritte)
}

// TestParseLebensmittelEingabe prüft die Freitext-Erfassung gegen den
// simulierten Server, inklusive Bereinigung (leere Namen, Anzahl-Default).
func TestParseLebensmittelEingabe(t *testing.T) {
	server := mockServer(t, map[string]any{"lebensmittel": []map[string]any{
		{"name": "Brot", "anzahl": 0, "einheit": "Stück", "gefrierfach": false},
		{"name": "  ", "anzahl": 2, "einheit": "Packung", "gefrierfach": true},
		{"name": "Erbsen", "anzahl": 1, "einheit": "Packung", "gefrierfach": true},
	}})
	defer server.Close()

	client := NewClient()
	eintraege, err := client.ParseLebensmittelEingabe("Füge Brot und Erbsen ins Gefrierfach hinzu")

	assert.NoError(t, err)
	assert.Len(t, eintraege, 2) // leerer Name wurde verworfen
	assert.Equal(t, "Brot", eintraege[0].Name)
	assert.Equal(t, uint16(1), eintraege[0].Anzahl) // Anzahl 0 -> 1
	assert.Equal(t, "Erbsen", eintraege[1].Name)
	assert.True(t, eintraege[1].Gefrierfach)
}

// TestKlassifiziereLebensmittel prüft die Batch-Klassifizierung inklusive
// Bereinigung ungültiger Werte.
func TestKlassifiziereLebensmittel(t *testing.T) {
	server := mockServer(t, map[string]any{"zuordnungen": []map[string]any{
		{"id": 1, "lebensmittel_kategorie": "Rindfleisch", "mahlzeiten": []string{"Mittagessen"}},
		{"id": 2, "lebensmittel_kategorie": "Ungültig", "mahlzeiten": []string{"Frühstück", "Brunch"}},
		{"id": 0, "lebensmittel_kategorie": "Vegan", "mahlzeiten": []string{}},
	}})
	defer server.Close()

	client := NewClient()
	zuordnungen, err := client.KlassifiziereLebensmittel([]KlassifiziereAnfrageEintrag{
		{ID: 1, Name: "Rinderhackfleisch"},
		{ID: 2, Name: "Kombucha"},
	})

	assert.NoError(t, err)
	assert.Len(t, zuordnungen, 2) // ID 0 wurde verworfen
	assert.Equal(t, "Rindfleisch", zuordnungen[1].LebensmittelKategorie)
	assert.Equal(t, []string{"Frühstück"}, zuordnungen[2].Mahlzeiten) // "Brunch" bereinigt
	assert.Equal(t, "", zuordnungen[2].LebensmittelKategorie)         // ungültig -> leer
}

// TestLeereKlassifizierung: ohne Einträge kein API-Aufruf.
func TestLeereKlassifizierung(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "k")
	client := NewClient()

	zuordnungen, err := client.KlassifiziereLebensmittel(nil)
	assert.NoError(t, err)
	assert.Empty(t, zuordnungen)
}

// TestKlassifizierPromptAuflistung prüft, dass alle Einträge mit ID im
// Prompt landen.
func TestKlassifizierPromptAuflistung(t *testing.T) {
	prompt := buildKlassifizierPrompt([]KlassifiziereAnfrageEintrag{
		{ID: 1, Name: "Kombucha"},
		{ID: 42, Name: "Yuzu-Paste"},
	})
	assert.Contains(t, prompt, "- 1 - Kombucha")
	assert.Contains(t, prompt, "- 42 - Yuzu-Paste")
	assert.Contains(t, prompt, "Küchenrichtung")
}
