// Prompts und Prompt-Aufbau für die LLM-Funktionen der App.
package llm

import (
	"fmt"
	"strings"

	"gitlab.com/ReneGumzej/kitchenops/pkg/models"
)

const systemPrompt = `Du bist ein erfahrener Küchenchef der App KitchenOps.
Deine Aufgabe: Erstelle schmackhafte, praktisch umsetzbare Rezepte auf Basis
des vorhandenen Inventars des Nutzers.

Regeln:
- Nutze bevorzugt Zutaten, die im Inventar vorhanden sind (vorhanden: true).
- Kleinere Grundzutaten (Salz, Pfeffer, Öl, Wasser) darfst du ergänzen, markiere
  sie mit vorhanden: false.
- Gib für jede Zutat eine realistische Menge und Einheit an.
- Beschreibe die Zubereitung in klaren, nummerierten Schritten (5-10 Schritte).
- Antworte ausschließlich auf Deutsch.
- Halte dich exakt an das vorgegebene JSON-Schema.`

const klassifizierSystemPrompt = `Du ordnest Lebensmittel einer Küchenverwaltung Kategorien zu.
Antworte ausschließlich mit dem vorgegebenen JSON-Schema.
Sei konservativ: bei Unsicherheit lieber leer lassen als falsch zuordnen.`

const erfassungSystemPrompt = `Du extrahierst Lebensmittel aus einem deutschen Freitext für eine Küchenverwaltung.

Regeln (strikt einhalten):
- Nenne jedes im Text erwähnte Lebensmittel genau einmal.
- Bei expliziten Mengen ("500 g Mehl", "zwei mal Sahne", "3 Eier", "eine Packung Erbsen") übernimm Anzahl und Einheit.
- WICHTIG: OHNE explizite Mengenangabe IMMER anzahl 1 und einheit "Stück" - auch wenn du eine typische Packungsgröße kennst. Nur was im Text steht, nichts dazu erfinden.
- Wenn der Text erwähnt, dass ein Lebensmittel ins Gefrierfach / in die Kühltruhe / in die Tiefkühlung gehört, setze gefrierfach auf true (sonst false).
- Normalisiere die Einheit auf die erlaubte Liste (Stück, g, kg, ml, l, Packung, Beutel, Dose, Glas, Flasche, Becher, Bund, Tafel). Schreibe "Stück" immer groß.
- Antworte ausschließlich mit dem vorgegebenen JSON-Schema.`

// buildUserPrompt baut die strukturierte Nutzer-Anfrage inklusive Inventar.
func buildUserPrompt(inventar []models.Lebensmittel, anfrage Anfrage) string {
	var sb strings.Builder

	sb.WriteString("Erstelle Rezeptideen für den Nutzer.\n\n")

	if anfrage.EssensKategorie != "" {
		sb.WriteString(fmt.Sprintf("Essenskategorie (verbindlich): %s\n", anfrage.EssensKategorie))
	}
	if anfrage.LebensmittelKategorie != "" {
		sb.WriteString(fmt.Sprintf("Lebensmittelkategorie (verbindlich): %s\n", anfrage.LebensmittelKategorie))
	}
	sb.WriteString(fmt.Sprintf("Anzahl Rezepte: %d\n", anfrage.Anzahl))

	if anfrage.Prompt != "" {
		sb.WriteString(fmt.Sprintf("\nZusätzlicher Wunsch des Nutzers:\n%s\n", anfrage.Prompt))
	}

	sb.WriteString("\nAktuelles Inventar (mit Kategorien):\n")
	if len(inventar) == 0 {
		sb.WriteString("(leer - schlage gängige Rezepte vor)")
	} else {
		for _, lm := range inventar {
			kategorie := "unbekannt"
			if lm.LebensmittelKategorie != nil {
				kategorie = string(*lm.LebensmittelKategorie)
			}
			mahlzeiten := ""
			if len(lm.Mahlzeiten) > 0 {
				teil := make([]string, len(lm.Mahlzeiten))
				for i, m := range lm.Mahlzeiten {
					teil[i] = string(m)
				}
				mahlzeiten = " [" + strings.Join(teil, ", ") + "]"
			}
			sb.WriteString(fmt.Sprintf("- %s: %d %s (Kategorie: %s%s)\n", lm.Name, lm.Anzahl, lm.Einheit, kategorie, mahlzeiten))
		}
	}

	return sb.String()
}

// buildKlassifizierPrompt baut die Nutzer-Anfrage für die Batch-Klassifizierung.
func buildKlassifizierPrompt(eintraege []KlassifiziereAnfrageEintrag) string {
	var sb strings.Builder
	sb.WriteString("Ordne folgende Lebensmittel einer Küchenrichtung und den passenden Mahlzeiten zu.\n\n")
	sb.WriteString("Küchenrichtung (genau eine, oder leer wenn unklar):\n")
	sb.WriteString("Pasta, Rindfleisch, Hühnerfleisch, Schweinefleisch, Vegetarisch, Vegan, Salat\n\n")
	sb.WriteString("Mahlzeiten (mehrere möglich): Frühstück, Mittagessen, Abendessen, Snack\n\n")
	sb.WriteString("Regeln:\n")
	sb.WriteString("- Nenne für jedes Lebensmittel die ID aus der Liste und die Zuordnung.\n")
	sb.WriteString("- Wenn du dir unsicher bist, lass die Küchenrichtung leer und die Mahlzeiten-Liste leer.\n")
	sb.WriteString("- Tierische Produkte wie Milch, Butter, Käse, Eier sind Vegetarisch (nicht Vegan).\n")
	sb.WriteString("- Pflanzliche Grundzutaten (Gemüse, Obst, Mehl, Nüsse) sind Vegetarisch oder Vegan.\n\n")
	sb.WriteString("Lebensmittel (id - name):\n")
	for _, e := range eintraege {
		sb.WriteString(fmt.Sprintf("- %d - %s\n", e.ID, e.Name))
	}
	return sb.String()
}

// bereinigeKlassifizierung validiert die LLM-Antwort gegen die erlaubten Werte.
// Ungültige Kategorien werden geleert.
func bereinigeKlassifizierung(zuordnungen []KlassifiziereEintrag) map[uint]KlassifiziereEintrag {
	erlaubteKategorien := map[string]bool{
		"Pasta": true, "Rindfleisch": true, "Hühnerfleisch": true, "Schweinefleisch": true,
		"Vegetarisch": true, "Vegan": true, "Salat": true,
	}
	erlaubteMahlzeiten := map[string]bool{
		"Frühstück": true, "Mittagessen": true, "Abendessen": true, "Snack": true,
	}

	result := make(map[uint]KlassifiziereEintrag, len(zuordnungen))
	for _, z := range zuordnungen {
		if z.ID == 0 {
			continue
		}
		if z.LebensmittelKategorie != "" && !erlaubteKategorien[z.LebensmittelKategorie] {
			z.LebensmittelKategorie = ""
		}
		gültigeMahlzeiten := make([]string, 0, len(z.Mahlzeiten))
		for _, m := range z.Mahlzeiten {
			if erlaubteMahlzeiten[m] {
				gültigeMahlzeiten = append(gültigeMahlzeiten, m)
			}
		}
		z.Mahlzeiten = gültigeMahlzeiten
		result[z.ID] = z
	}
	return result
}
