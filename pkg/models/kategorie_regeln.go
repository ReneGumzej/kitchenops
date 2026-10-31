package models

import "strings"

// ============================================================================
// Regel-Klassifikator: ordnet Lebensmitteln anhand ihres Namens schnell und
// ohne API-Aufruf eine Küchenrichtung und passende Mahlzeiten zu.
// Läuft rein in-memory (O(Artikel × Schlüsselwörter)) und wird automatisch
// beim Anlegen/Bearbeiten angewendet, wenn keine Kategorie gesetzt ist.
// ============================================================================

type kategorieRegel struct {
	schluesselwörter []string
	kategorie        LebensmittelKategorie
	mahlzeiten       []EssensKategorie
}

// kategorieRegeln: geordnet nach Spezifität - der erste Treffer gewinnt.
var kategorieRegeln = []kategorieRegel{
	// --- Fleisch (spezifisch zuerst) ---
	{[]string{"rinderhack", "rindfleisch", "rindersteak", "hackfleisch vom rind", "rosbeef", "rinderbraten"}, LebensmittelKategorieRindfleisch, nil},
	{[]string{"hähnchen", "hühner", "hähnchenbrust", "hühnerbrust", "hühnchen", "putenbrust", "pute", "hühnerei weiß"}, LebensmittelKategorieHuehnerfleisch, nil},
	{[]string{"schwein", "speck", "bacon", "schinken", "schweinefilet", "kassler", "mettwurst", "salami"}, LebensmittelKategorieSchweinefleisch, nil},

	// --- Pasta ---
	{[]string{"nudel", "pasta", "spaghetti", "macaroni", "maccheroni", "penne", "fusilli", "farfalle", "lasagne", "tagliatelle", "gnocchi", "maultasche"}, LebensmittelKategoriePasta, nil},

	// --- Salat ---
	{[]string{"salat", "gurke", "kopfsalat", "rucola", "eisberg", "romana", "feldsalat", "spinat", "lattugh"}, LebensmittelKategorieSalat, nil},

	// --- Vegan (pflanzlich, ohne tierische Produkte) ---
	{[]string{"tofu", "seitan", "tempeh", "sojadrink", "haferdrink", "mandeldrink", "vegane"}, LebensmittelKategorieVegan, nil},

	// --- Vegetarisch (mit Eiern/Milchprodukten) ---
	{[]string{"eier", "ei ", " ei", "milch", "butter", "käse", "joghurt", "quark", "schmand", "frischkäse", "mozzarella", "feta", "parmesan", "gouda", "emmentaler", "cream cheese", "sahne"}, LebensmittelKategorieVegetarisch, nil},

	// --- Weitere pflanzliche Grundzutaten: Vegetarisch als Basis ---
	{[]string{"mehl", "zucker", "reis", "kartoffel", "zwiebel", "knoblauch", "tomate", "paprika", "karotte", "möhre", "zucchini", "aubergine", "brokkoli", "blumenkohl", "bohne", "linse", "kichererbse", "erbs", "mais", "pilz", "champignon", "avocado", "banane", "apfel", "birne", "orange", "zitrone", "beere", "himbeer", "erdbeer", "heidelbeer", "haferflocken", "müsli", "brot", "brötchen", "toast", "öl", "essig", "honig", "marmelade", "nuss", "mandel", "walnuss", "erdnuss", "hafer", "gerste", "grünkern", "hirse", "quinoa", "couscous", "bulgur", "linsen"}, LebensmittelKategorieVegetarisch, nil},
}

// mahlzeitRegeln: typische Frühstücks-/Snack-Artikel.
var fruehstueckSchluesselwoerter = []string{
	"eier", "milch", "butter", "marmelade", "honig", "müsli", "haferflocken",
	"brötchen", "brot", "toast", "käse", "joghurt", "quark", "frischkäse",
	"cornflakes", "kaffee", "tee", "saft", "orangensaft", "apfelsaft", "banane",
	"erdbeere", "himbeere", "heidelbeere", "birne", "apfel",
}

var snackSchluesselwoerter = []string{
	"chips", "nuts", "nuss", "mandel", "erdnuss", "walnuss", "schokolade",
	"kekse", "keks", "cracker", "salzstangen", "popcorn", "studentenfutter",
	"obst", "apfel", "banane", "birne", "trauben", "müsli", "joghurt",
}

// KlassifiziereRegeln bestimmt Küchenrichtung und Mahlzeiten anhand des
// Namens. Gibt nil zurück, wenn keine Regel sicher trifft.
func KlassifiziereRegeln(name string) (*LebensmittelKategorie, []EssensKategorie) {
	nameLower := " " + strings.ToLower(strings.TrimSpace(name)) + " "

	var kategorie *LebensmittelKategorie
	for _, regel := range kategorieRegeln {
		for _, sw := range regel.schluesselwörter {
			if strings.Contains(nameLower, sw) {
				k := regel.kategorie
				kategorie = &k
				break
			}
		}
		if kategorie != nil {
			break
		}
	}

	var mahlzeiten []EssensKategorie
	seen := make(map[EssensKategorie]bool)
	addMahlzeit := func(k EssensKategorie) {
		if !seen[k] {
			seen[k] = true
			mahlzeiten = append(mahlzeiten, k)
		}
	}
	for _, sw := range fruehstueckSchluesselwoerter {
		if strings.Contains(nameLower, sw) {
			addMahlzeit(EssensKategorieFruehstueck)
			break
		}
	}
	for _, sw := range snackSchluesselwoerter {
		if strings.Contains(nameLower, sw) {
			addMahlzeit(EssensKategorieSnack)
			break
		}
	}
	// Ohne spezifische Mahlzeit passen die meisten Zutaten flexibel
	if len(mahlzeiten) == 0 && kategorie != nil {
		addMahlzeit(EssensKategorieMittagessen)
		addMahlzeit(EssensKategorieAbendessen)
	}

	return kategorie, mahlzeiten
}

// IstFleisch prüft, ob die Kategorie eine Fleisch-Kategorie ist.
func (k LebensmittelKategorie) IstFleisch() bool {
	switch k {
	case LebensmittelKategorieRindfleisch, LebensmittelKategorieHuehnerfleisch, LebensmittelKategorieSchweinefleisch:
		return true
	}
	return false
}
