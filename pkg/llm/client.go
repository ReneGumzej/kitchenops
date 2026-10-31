// Package llm stellt den provider-agnostischen LLM-Client der App bereit.
// Basis ist das GoAI-Framework (github.com/zendev-sh/goai): die App spricht
// alle Anbieter über dieselbe Kern-API (GenerateObject für strukturierte
// Ausgaben) an.
//
// Unterstützte Anbieter und ihre API-Key-Umgebungsvariablen:
//
//	mistral    MISTRAL_API_KEY                (Standard)
//	openai     OPENAI_API_KEY
//	anthropic  ANTHROPIC_API_KEY
//	google     GOOGLE_GENERATIVE_AI_API_KEY oder GEMINI_API_KEY
//
// Konfiguration:
//
//	LLM_PROVIDER  mistral | openai | anthropic | google
//	              (ohne Angabe: Anbieter wird am vorhandenen API-Key erkannt)
//	LLM_MODEL     Modellname des Anbieters
//	              (Default je Anbieter, z.B. mistral-small-2603)
package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/anthropic"
	"github.com/zendev-sh/goai/provider/google"
	"github.com/zendev-sh/goai/provider/mistral"
	"github.com/zendev-sh/goai/provider/openai"

	"gitlab.com/ReneGumzej/kitchenops/pkg/logging"
	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
)

// Anbieter-Bezeichner
const (
	ProviderMistral   = "mistral"
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGoogle    = "google"
)

// defaultModelle je Anbieter (überschreibbar über LLM_MODEL bzw.
// MISTRAL_MODEL für Mistral aus Abwärtskompatibilität).
var defaultModelle = map[string]string{
	ProviderMistral:   "mistral-small-2603",
	ProviderOpenAI:    "gpt-4o",
	ProviderAnthropic: "claude-sonnet-4-5",
	ProviderGoogle:    "gemini-2.5-flash",
}

// apiKeyEnv ist die Umgebungsvariable je Anbieter.
func apiKeyEnv(anbieter string) string {
	switch anbieter {
	case ProviderMistral:
		return "MISTRAL_API_KEY"
	case ProviderOpenAI:
		return "OPENAI_API_KEY"
	case ProviderAnthropic:
		return "ANTHROPIC_API_KEY"
	case ProviderGoogle:
		if os.Getenv("GOOGLE_GENERATIVE_AI_API_KEY") != "" {
			return "GOOGLE_GENERATIVE_AI_API_KEY"
		}
		return "GEMINI_API_KEY"
	}
	return ""
}

// Client ist der provider-agnostische LLM-Client der App.
type Client struct {
	anbieter  string
	modell    string
	verfügbar bool
	model     provider.LanguageModel
}

// NewClient erstellt den Client aus den Umgebungsvariablen.
// Ohne explizites LLM_PROVIDER wird der Anbieter am vorhandenen API-Key
// erkannt - wer nur einen Key setzt, bekommt automatisch den passenden
// Anbieter mit seinem Standardmodell. Ohne jeden Key bleibt der Client
// nicht verfügbar (Verfügbar() == false).
func NewClient() *Client {
	anbieter := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	if anbieter == "" {
		anbieter = erkenneAnbieter()
	}

	modell := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if modell == "" && anbieter == ProviderMistral {
		// Abwärtskompatibilität zur bisherigen Konfiguration
		modell = strings.TrimSpace(os.Getenv("MISTRAL_MODEL"))
	}
	if modell == "" {
		modell = defaultModelle[anbieter]
	}

	key := os.Getenv(apiKeyEnv(anbieter))

	switch anbieter {
	case ProviderMistral:
		opts := []mistral.Option{}
		if key != "" {
			opts = append(opts, mistral.WithAPIKey(key))
		}
		return &Client{anbieter, modell, key != "", mistral.Chat(modell, opts...)}
	case ProviderOpenAI:
		opts := []openai.Option{}
		if key != "" {
			opts = append(opts, openai.WithAPIKey(key))
		}
		return &Client{anbieter, modell, key != "", openai.Chat(modell, opts...)}
	case ProviderAnthropic:
		opts := []anthropic.Option{}
		if key != "" {
			opts = append(opts, anthropic.WithAPIKey(key))
		}
		return &Client{anbieter, modell, key != "", anthropic.Chat(modell, opts...)}
	case ProviderGoogle:
		opts := []google.Option{}
		if key != "" {
			opts = append(opts, google.WithAPIKey(key))
		}
		return &Client{anbieter, modell, key != "", google.Chat(modell, opts...)}
	}

	logging.Warn("unbekannter LLM_PROVIDER", "anbieter", anbieter,
		"erlaubt", "mistral, openai, anthropic, google")
	return &Client{anbieter: anbieter, modell: modell, verfügbar: false}
}

// erkenneAnbieter bestimmt den Anbieter am ersten vorhandenen API-Key.
// Mistral hat Vorrang, da es der bisherige Standard der App ist.
func erkenneAnbieter() string {
	for _, anbieter := range []string{ProviderMistral, ProviderOpenAI, ProviderAnthropic, ProviderGoogle} {
		if os.Getenv(apiKeyEnv(anbieter)) != "" {
			return anbieter
		}
	}
	return ProviderMistral
}

// Verfügbar meldet, ob ein API-Key für den Anbieter konfiguriert ist.
func (c *Client) Verfügbar() bool {
	return c.verfügbar
}

// Model gibt den konfigurierten Modellnamen zurück.
func (c *Client) Model() string {
	return c.modell
}

// Anbieter gibt den aktiven Anbieter zurück (mistral, openai, ...).
func (c *Client) Anbieter() string {
	return c.anbieter
}

// =============================================================================
// Aufruf-Hilfen
// =============================================================================

// aufruf führt ein GenerateObject mit System-Prompt, Nutzer-Prompt und
// Temperatur aus (nur wenn das Modell Temperatur unterstützt) und
// protokolliert Modell, Dauer und Token-Verbrauch.
func aufruf[T any](funktion string, c *Client, nutzerPrompt string, temperatur float64) (T, error) {
	var leer T

	opts := []goai.Option{
		goai.WithSystem(systemPromptFunktion(funktion)),
		goai.WithPrompt(nutzerPrompt),
	}
	if provider.ModelCapabilitiesOf(c.model).Temperature {
		opts = append(opts, goai.WithTemperature(temperatur))
	}

	start := time.Now()
	result, err := goai.GenerateObject[T](context.Background(), c.model, opts...)
	if err != nil {
		logging.Error("llm-aufruf fehlgeschlagen",
			"funktion", funktion,
			"anbieter", c.anbieter,
			"modell", c.modell,
			"fehler", err.Error())
		return leer, err
	}

	logging.Info("llm-aufruf",
		"funktion", funktion,
		"anbieter", c.anbieter,
		"modell", c.modell,
		"dauer_ms", time.Since(start).Milliseconds(),
		"prompt_tokens", result.Usage.InputTokens,
		"completion_tokens", result.Usage.OutputTokens,
	)
	return result.Object, nil
}

// systemPromptFunktion liefert den System-Prompt je LLM-Funktion.
func systemPromptFunktion(funktion string) string {
	switch funktion {
	case "rezept-generierung":
		return systemPrompt
	case "batch-klassifizierung":
		return klassifizierSystemPrompt
	case "text-erfassung":
		return erfassungSystemPrompt
	}
	return systemPrompt
}

// =============================================================================
// Öffentliche LLM-Funktionen
// =============================================================================

// GeneriereRezepte schickt Inventar + strukturierte Vorauswahl an das LLM
// und liefert die strukturiert geparsten Rezeptvorschläge zurück.
func (c *Client) GeneriereRezepte(inventar []models.Lebensmittel, anfrage Anfrage) ([]Rezept, error) {
	if !c.verfügbar {
		return nil, fmt.Errorf("kein API-Key für Anbieter %q konfiguriert", c.anbieter)
	}
	if anfrage.Anzahl == 0 {
		anfrage.Anzahl = 1
	}

	antwort, err := aufruf[RezeptAntwort]("rezept-generierung", c, buildUserPrompt(inventar, anfrage), 0.7)
	if err != nil {
		return nil, err
	}
	if len(antwort.Rezepte) == 0 {
		return nil, fmt.Errorf("das Modell hat keine passenden Rezepte gefunden")
	}
	return antwort.Rezepte, nil
}

// KlassifiziereLebensmittel klassifiziert eine Liste von Lebensmitteln in
// EINEM einzigen API-Aufruf (performant auch bei großem Inventar).
// Die Antwort wird gegen die erlaubten Kategorien validiert; ungültige
// Werte werden zu "keine Zuordnung" bereinigt.
func (c *Client) KlassifiziereLebensmittel(eintraege []KlassifiziereAnfrageEintrag) (map[uint]KlassifiziereEintrag, error) {
	if !c.verfügbar {
		return nil, fmt.Errorf("kein API-Key für Anbieter %q konfiguriert", c.anbieter)
	}
	if len(eintraege) == 0 {
		return map[uint]KlassifiziereEintrag{}, nil
	}

	antwort, err := aufruf[KlassifiziereAntwort]("batch-klassifizierung", c, buildKlassifizierPrompt(eintraege), 0.1)
	if err != nil {
		return nil, err
	}
	return bereinigeKlassifizierung(antwort.Zuordnungen), nil
}

// ParseLebensmittelEingabe extrahiert aus einem freien deutschen Text
// strukturierte Lebensmittel-Einträge.
func (c *Client) ParseLebensmittelEingabe(text string) ([]ErfassungEintrag, error) {
	if !c.verfügbar {
		return nil, fmt.Errorf("kein API-Key für Anbieter %q konfiguriert", c.anbieter)
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("leere Eingabe")
	}

	antwort, err := aufruf[ErfassungAntwort]("text-erfassung", c, text, 0.1)
	if err != nil {
		return nil, err
	}

	// Bereinigen: leere Namen verwerfen, Anzahl 0 -> 1
	gültig := make([]ErfassungEintrag, 0, len(antwort.Lebensmittel))
	for _, e := range antwort.Lebensmittel {
		if strings.TrimSpace(e.Name) == "" {
			continue
		}
		e.Name = strings.TrimSpace(e.Name)
		if e.Anzahl == 0 {
			e.Anzahl = 1
		}
		if e.Einheit == "" {
			e.Einheit = "Stück"
		}
		gültig = append(gültig, e)
	}
	return gültig, nil
}
