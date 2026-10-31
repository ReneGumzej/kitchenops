/**
 * KitchenOps - Rezepte-Seite
 * KI-gestützte Rezept-Generierung über das Backend (mistral-small-2603)
 * mit strukturierte Vorauswahl (Kategorien) und freiem Wunsch-Text.
 */

const auswahl = {
    essensKategorie: '',
    lebensmittelKategorie: '',
    anzahl: 1
};

let generierteRezepte = [];
let gespeicherteListe = [];
// Namen aller offenen Einkaufslisten-Artikel (kleingeschrieben).
// Dient dazu, den Haken-Status der Zutat-Buttons beim Rendern korrekt
// wiederherzustellen - auch nach Seiten-Reload oder Modal-Neuöffnung.
let einkaufslisteOffen = new Set();

function PageInit() {
    bindChips();
    bindGenerator();
    bindRezeptCardButtons();

    DataManager.loadLebensmittel(true).then(() => {
        updateInventarHint();
        Store.subscribe(() => updateInventarHint());
    });
    loadGespeicherte();
    loadEinkaufsliste();
}

// Lädt die offenen Einkaufslisten-Artikel, um den Button-Zustand
// ("bereits auf der Liste") beim Rendern korrekt abzubilden.
async function loadEinkaufsliste() {
    try {
        const data = await api.getAllEinkaufsartikel();
        const artikel = Array.isArray(data?.artikel) ? data.artikel : [];
        einkaufslisteOffen = new Set(
            artikel.filter(a => !a.erledigt)
                .map(a => (a.name || '').toLowerCase())
                .filter(Boolean));
    } catch (error) {
        einkaufslisteOffen = new Set();
    }
}

// istAufEinkaufsliste prüft, ob ein Artikel (Name, groß/klein egal) offen
// auf der Einkaufsliste steht.
function istAufEinkaufsliste(name) {
    return einkaufslisteOffen.has((name || '').toLowerCase());
}

// -------------------------------------------------------------------------
// Einkaufsliste-Übernahme fehlender Zutaten
// -------------------------------------------------------------------------
function bindRezeptCardButtons() {
    // Ein delegierter Listener je Container - gilt für KI-Ergebnisse
    // UND für die Detail-Modals der gespeicherten Rezepte
    const handler = onRezeptCardClick;
    document.getElementById('rezeptErgebnisse').addEventListener('click', handler);
    document.getElementById('modalBody').addEventListener('click', handler);
}

function onRezeptCardClick(e) {
    const saveBtn = e.target.closest('.rezept-save-btn');
    const alleBtn = e.target.closest('.alle-einkauf-btn');
    const zutatBtn = e.target.closest('.zutat-einkauf-btn');

    if (saveBtn) {
        e.stopPropagation();
        speichern(parseInt(saveBtn.dataset.index, 10));
    } else if (alleBtn) {
        e.stopPropagation();
        alleFehlendeZurEinkaufsliste(alleBtn);
    } else if (zutatBtn) {
        e.stopPropagation();
        zutatZurEinkaufsliste(zutatBtn);
    }
}

// parseMenge wandelt die LLM-Mengenangabe in eine Einkaufslisten-Anzahl um.
// Bruchzahlen werden aufgerundet (0.5 Bund -> 1), Ungültiges ergibt 1.
function parseMenge(menge) {
    const n = parseFloat(String(menge).replace(',', '.'));
    if (!Number.isFinite(n) || n <= 0) return 1;
    return Math.max(1, Math.ceil(n));
}

// grundNahrungsmittel liefert Menge und Einheit für die Einkaufsliste.
// Bewusst NICHT die Rezeptmengen (z.B. "1 Prise"), sondern das grundlegende
// Lebensmittel: Einheit aus dem Inventar (falls vorhanden), sonst "Stück";
// bei g/ml ein sinnvoller Grundbedarf von 500, sonst 1.
function grundNahrungsmittel(name) {
    const { lebensmittel } = Store.getState();
    const bekannt = lebensmittel.find(
        l => l.name.toLowerCase() === (name || '').toLowerCase());

    const einheit = bekannt ? bekannt.einheit : 'Stück';
    const grundbedarf = { 'g': 500, 'ml': 500 };
    const anzahl = grundbedarf[(einheit || '').toLowerCase()] || 1;

    return { anzahl, einheit: einheit || 'Stück' };
}

// getRezeptFromBtn liefert das Rezept-Objekt hinter einem Karten-Button.
function getRezeptFromBtn(btn) {
    const index = parseInt(btn.dataset.index, 10);
    if (btn.dataset.source === 'saved') {
        return gespeicherteListe[index];
    }
    return generierteRezepte[index];
}

// Option 2: einzelne fehlende Zutat auf die Einkaufsliste
async function zutatZurEinkaufsliste(btn) {
    const rezept = getRezeptFromBtn(btn);
    const name = btn.dataset.name;

    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i>';

    try {
        const { anzahl, einheit } = grundNahrungsmittel(name);
        await api.createEinkaufsartikel({
            name: name,
            anzahl: anzahl,
            einheit: einheit,
            notiz: rezept ? `für: ${rezept.name}` : ''
        });
        // Zustand lokal merken, damit der Haken auch nach Neu-Rendern bleibt
        einkaufslisteOffen.add(name.toLowerCase());
        Toast.success(`"${name}" ist auf der Einkaufsliste (${anzahl} ${einheit})`);
        btn.innerHTML = '<i class="fas fa-check"></i>';
        btn.classList.add('hinzugefuegt');
        btn.title = 'Bereits auf der Einkaufsliste';
    } catch (error) {
        Toast.error(`Fehler beim Hinzufügen: ${error.message}`);
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-cart-plus"></i>';
    }
}

// Option 1: alle fehlenden Zutaten eines Rezepts auf einmal
async function alleFehlendeZurEinkaufsliste(btn) {
    const rezept = getRezeptFromBtn(btn);
    if (!rezept) return;

    const fehlende = (rezept.zutaten || []).filter(z => !z.vorhanden);
    if (fehlende.length === 0) {
        Toast.info('Alle Zutaten sind bereits im Inventar');
        return;
    }

    const original = btn.innerHTML;
    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Füge hinzu...';

    try {
        for (const z of fehlende) {
            const { anzahl, einheit } = grundNahrungsmittel(z.name);
            await api.createEinkaufsartikel({
                name: z.name,
                anzahl: anzahl,
                einheit: einheit,
                notiz: `für: ${rezept.name}`
            });
            einkaufslisteOffen.add(z.name.toLowerCase());
        }
        Toast.success(`${fehlende.length} Zutaten sind auf der Einkaufsliste`);
        btn.innerHTML = '<i class="fas fa-check"></i> Auf der Einkaufsliste';
        btn.classList.add('hinzugefuegt');

        // Einzel-Buttons ebenfalls als erledigt markieren
        const karte = btn.closest('.rezept-card');
        karte?.querySelectorAll('.zutat-einkauf-btn').forEach(b => {
            b.disabled = true;
            b.innerHTML = '<i class="fas fa-check"></i>';
            b.classList.add('hinzugefuegt');
        });
    } catch (error) {
        Toast.error(`Fehler beim Hinzufügen: ${error.message}`);
        btn.disabled = false;
        btn.innerHTML = original;
    }
}

// -------------------------------------------------------------------------
// Auswahl-Buttons
// -------------------------------------------------------------------------
function bindChips() {
    document.querySelectorAll('#essensKategorieChips .choice-chip').forEach(chip => {
        chip.addEventListener('click', () => {
            const active = chip.classList.contains('active');
            document.querySelectorAll('#essensKategorieChips .choice-chip')
                .forEach(c => c.classList.remove('active'));
            auswahl.essensKategorie = active ? '' : chip.dataset.value;
            if (!active) chip.classList.add('active');
        });
    });

    document.querySelectorAll('#lebensmittelKategorieChips .choice-chip').forEach(chip => {
        chip.addEventListener('click', () => {
            const active = chip.classList.contains('active');
            document.querySelectorAll('#lebensmittelKategorieChips .choice-chip')
                .forEach(c => c.classList.remove('active'));
            auswahl.lebensmittelKategorie = active ? '' : chip.dataset.value;
            if (!active) chip.classList.add('active');
        });
    });

    document.querySelectorAll('#anzahlChips .choice-chip').forEach(chip => {
        chip.addEventListener('click', () => {
            document.querySelectorAll('#anzahlChips .choice-chip')
                .forEach(c => c.classList.remove('active'));
            chip.classList.add('active');
            auswahl.anzahl = parseInt(chip.dataset.value, 10);
        });
    });

    // Default: 1 Rezept
    document.querySelector('#anzahlChips .choice-chip[data-value="1"]').classList.add('active');
}

function bindGenerator() {
    document.getElementById('generateBtn').addEventListener('click', generieren);
}

function updateInventarHint() {
    const { lebensmittel } = Store.getState();
    const hint = document.getElementById('inventarHint');
    if (!hint) return;
    hint.innerHTML = lebensmittel.length > 0
        ? `<i class="fas fa-box-open"></i> ${lebensmittel.length} Lebensmittel im Inventar verfügbar`
        : `<i class="fas fa-triangle-exclamation"></i> Inventar ist leer - lege zuerst Lebensmittel an`;
}

// -------------------------------------------------------------------------
// Generierung
// -------------------------------------------------------------------------
async function generieren() {
    const prompt = document.getElementById('rezeptPrompt').value.trim();

    if (!prompt && !auswahl.essensKategorie && !auswahl.lebensmittelKategorie) {
        Toast.warning('Bitte wählen Sie eine Kategorie oder geben Sie einen Wunsch-Text ein');
        return;
    }

    const btn = document.getElementById('generateBtn');
    const ergebnisse = document.getElementById('rezeptErgebnisse');
    const kiLoading = document.getElementById('kiLoading');

    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Generiere...';
    kiLoading.classList.remove('hidden');
    ergebnisse.innerHTML = '';
    generierteRezepte = [];

    try {
        const data = await api.generateRezepte({
            prompt: prompt,
            essens_kategorie: auswahl.essensKategorie,
            lebensmittel_kategorie: auswahl.lebensmittelKategorie,
            anzahl: auswahl.anzahl
        });

        generierteRezepte = Array.isArray(data?.rezepte) ? data.rezepte : [];
        renderErgebnisse();
    } catch (error) {
        ergebnisse.innerHTML = `
            <div class="ki-error">
                <i class="fas fa-circle-exclamation"></i>
                <h3>Generierung fehlgeschlagen</h3>
                <p>${UI.escapeHtml(error.message)}</p>
            </div>`;
        Toast.error(error.message);
    } finally {
        kiLoading.classList.add('hidden');
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-wand-magic-sparkles"></i> Rezepte generieren';
    }
}

// -------------------------------------------------------------------------
// Rendering: generierte Rezepte
// -------------------------------------------------------------------------
function renderErgebnisse() {
    const container = document.getElementById('rezeptErgebnisse');

    if (generierteRezepte.length === 0) {
        container.innerHTML = UI.emptyState(
            'fa-lightbulb',
            'Keine Rezepte gefunden',
            'Versuchen Sie eine andere Auswahl oder formulieren Sie den Wunsch anders.'
        );
        return;
    }

    container.innerHTML = `
        <h2 class="mb-md"><i class="fas fa-wand-magic-sparkles text-primary"></i> Vorschläge der Küchen-KI</h2>
        <div class="rezept-grid">
            ${generierteRezepte.map((r, i) => rezeptCard(r, i, { source: 'generated' })).join('')}
        </div>
    `;
}

function rezeptCard(rezept, index = null, opts = {}) {
    const schwierigkeitBadge = schwierigkeitClass(rezept.schwierigkeit);
    const showSave = opts.saveButton !== false;
    const source = opts.source || 'generated';

    // Fallbacks für alte gespeicherte Einträge ohne Details
    const beschreibung = rezept.beschreibung ||
        'Für dieses Rezept ist keine Beschreibung gespeichert.';
    const schritte = rezept.schritte || [];
    const fehlende = (rezept.zutaten || []).filter(z => !z.vorhanden);
    const fehlendeOffen = fehlende.filter(z => !istAufEinkaufsliste(z.name));

    return `
        <article class="rezept-card">
            <div class="rezept-card-header">
                <div class="rezept-icon"><i class="fas fa-utensils"></i></div>
                <div class="rezept-titel">
                    <h3>${UI.escapeHtml(rezept.name)}</h3>
                    <p>${UI.escapeHtml(beschreibung)}</p>
                </div>
            </div>

            <div class="rezept-badges">
                ${kategorieBadge(rezept.essens_kategorie, 'essens')}
                ${kategorieBadge(rezept.lebensmittel_kategorie, 'lebensmittel')}
                ${rezept.zubereitungszeit_minuten ? `<span class="rezept-meta-chip"><i class="fas fa-clock"></i> ${rezept.zubereitungszeit_minuten} min</span>` : ''}
                ${rezept.schwierigkeit ? `<span class="rezept-meta-chip"><i class="fas fa-gauge-high"></i> <span class="badge ${schwierigkeitBadge}">${UI.escapeHtml(rezept.schwierigkeit)}</span></span>` : ''}
                ${rezept.portionen ? `<span class="rezept-meta-chip"><i class="fas fa-user-group"></i> ${rezept.portionen} Port.</span>` : ''}
                ${opts.gespeichertAm ? `<span class="rezept-meta-chip"><i class="fas fa-calendar"></i> ${UI.formatDate(opts.gespeichertAm)}</span>` : ''}
            </div>

            <div class="rezept-sectionen">
                <div class="rezept-section">
                    <h4><i class="fas fa-basket-shopping"></i> Zutaten</h4>
                    ${(rezept.zutaten || []).length > 0 ? `
                        <ul class="zutaten-liste">
                            ${(rezept.zutaten || []).map(z => `
                                <li class="${z.vorhanden ? 'vorhanden' : 'fehlt'}">
                                    <i class="fas ${z.vorhanden ? 'fa-circle-check' : 'fa-circle-xmark'}"></i>
                                    <span>${UI.escapeHtml(z.name)}</span>
                                    ${!z.vorhanden ? `
                                        <button type="button" class="zutat-einkauf-btn${istAufEinkaufsliste(z.name) ? ' hinzugefuegt' : ''}"
                                                ${istAufEinkaufsliste(z.name) ? 'disabled title="Bereits auf der Einkaufsliste"' : 'title="Diese Zutat zur Einkaufsliste hinzufügen"'}
                                                data-source="${source}" data-index="${index}"
                                                data-name="${UI.escapeHtml(z.name)}"
                                                data-menge="${UI.escapeHtml(z.menge || '')}"
                                                data-einheit="${UI.escapeHtml(z.einheit || '')}">
                                            <i class="fas ${istAufEinkaufsliste(z.name) ? 'fa-check' : 'fa-cart-plus'}"></i>
                                        </button>` : ''}
                                    <span class="zutat-menge">${z.menge ? `${UI.escapeHtml(z.menge)} ${UI.escapeHtml(z.einheit)}` : UI.escapeHtml(z.einheit || '')}</span>
                                </li>`).join('')}
                        </ul>
                        ${fehlende.length > 0 ? `
                            <p class="einkauf-hinweis text-muted text-sm">
                                <i class="fas fa-circle-info"></i>
                                Fehlende Zutaten: einzeln per <i class="fas fa-cart-plus"></i> oder alle zusammen unten hinzufügen.
                            </p>` : ''}
                    ` : '<p class="text-muted text-sm">Keine Zutaten verknüpft.</p>'}
                </div>

                <div class="rezept-section">
                    <h4><i class="fas fa-list-ol"></i> Zubereitung</h4>
                    ${schritte.length > 0 ? `
                        <ol class="schritte-liste">
                            ${schritte.map(s => `<li>${UI.escapeHtml(s)}</li>`).join('')}
                        </ol>` : '<p class="text-muted text-sm">Keine Zubereitungsschritte gespeichert.</p>'}
                </div>
            </div>

            ${(fehlende.length > 0 || showSave) ? `
                <div class="rezept-card-footer">
                    ${fehlende.length > 0 ? (fehlendeOffen.length > 0 ? `
                        <button type="button" class="btn btn-secondary alle-einkauf-btn"
                                title="Alle fehlenden Zutaten zur Einkaufsliste hinzufügen"
                                data-source="${source}" data-index="${index}">
                            <i class="fas fa-cart-plus"></i> ${fehlendeOffen.length} fehlende ${fehlendeOffen.length === 1 ? 'Zutat' : 'Zutaten'} einkaufen
                        </button>` : `
                        <button type="button" class="btn btn-outline alle-einkauf-btn" disabled title="Alle fehlenden Zutaten sind bereits auf der Einkaufsliste">
                            <i class="fas fa-check"></i> Alle auf der Einkaufsliste
                        </button>`) : ''}
                    ${showSave ? `
                        <button type="button" class="btn btn-primary rezept-save-btn" data-index="${index}">
                            <i class="fas fa-bookmark"></i> Rezept speichern
                        </button>` : ''}
                </div>` : ''}
        </article>
    `;
}

function kategorieBadge(name, typ) {
    if (!name) return '';
    const klasse = typ === 'essens' ? 'badge-info' : 'badge-success';
    const icon = typ === 'essens' ? 'fa-utensils' : 'fa-leaf';
    return `<span class="badge ${klasse}"><i class="fas ${icon}"></i> ${UI.escapeHtml(name)}</span>`;
}

function schwierigkeitClass(schwierigkeit) {
    switch ((schwierigkeit || '').toLowerCase()) {
        case 'einfach': return 'badge-success';
        case 'schwer': return 'badge-error';
        default: return 'badge-warning';
    }
}

// -------------------------------------------------------------------------
// Speichern
// -------------------------------------------------------------------------
async function speichern(index) {
    const rezept = generierteRezepte[index];
    if (!rezept) return;

    // Zutaten-Namen auf Inventar-IDs mappen (nur vorhandene Zutaten)
    const { lebensmittel } = Store.getState();
    const zutatenIDs = [];
    for (const z of (rezept.zutaten || [])) {
        if (!z.vorhanden) continue;
        const match = lebensmittel.find(
            l => l.name.toLowerCase() === z.name.toLowerCase());
        if (match && !zutatenIDs.includes(match.id)) {
            zutatenIDs.push(match.id);
        }
    }

    const btn = document.querySelector(`.rezept-save-btn[data-index="${index}"]`);
    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Speichere...';

    try {
        const result = await api.saveRezept({
            name: rezept.name,
            lebensmittel_kategorie: rezept.lebensmittel_kategorie,
            essens_kategorie: rezept.essens_kategorie,
            beschreibung: rezept.beschreibung,
            zutaten: (rezept.zutaten || []).map(z => ({
                name: z.name,
                menge: z.menge,
                einheit: z.einheit,
                vorhanden: !!z.vorhanden
            })),
            schritte: rezept.schritte || [],
            zubereitungszeit_minuten: rezept.zubereitungszeit_minuten || 0,
            schwierigkeit: rezept.schwierigkeit || '',
            portionen: rezept.portionen || 0,
            zutaten_ids: zutatenIDs
        });
        Toast.success(result?.message || 'Rezept gespeichert');
        btn.innerHTML = '<i class="fas fa-check"></i> Gespeichert';
        await loadGespeicherte();
    } catch (error) {
        Toast.error(`Fehler beim Speichern: ${error.message}`);
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-bookmark"></i> Rezept speichern';
    }
}

// -------------------------------------------------------------------------
// Gespeicherte Rezepte
// -------------------------------------------------------------------------
async function loadGespeicherte() {
    try {
        const data = await api.getAllRezepte();
        gespeicherteListe = Array.isArray(data?.rezepte) ? data.rezepte : [];
        renderGespeicherte();
    } catch (error) {
        // Stiller Fehler: gespeicherte Rezepte sind nur Zusatzfunktion
        console.warn('Gespeicherte Rezepte konnten nicht geladen werden:', error);
        gespeicherteListe = [];
        renderGespeicherte();
    }
}

// Löscht ein gespeichertes Rezept nach Bestätigung
async function deleteRezept(id) {
    const rezept = gespeicherteListe.find(r => r.id === id);
    const name = rezept ? rezept.name : 'Rezept';

    if (!await UI.confirm(`Möchten Sie das Rezept "${name}" wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.`)) {
        return;
    }

    const btn = document.querySelector(`[data-id="${id}"]`);
    if (btn) {
        btn.disabled = true;
        btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Lösche...';
    }

    try {
        const result = await api.deleteRezept(id);
        Toast.success(result?.message || 'Rezept erfolgreich gelöscht');
        Modal.close();
        await loadGespeicherte();
    } catch (error) {
        Toast.error(`Fehler beim Löschen: ${error.message}`);
        if (btn) {
            btn.disabled = false;
            btn.innerHTML = '<i class="fas fa-trash"></i>';
        }
    }
}

function renderGespeicherte() {
    const container = document.getElementById('gespeicherteRezepte');

    if (gespeicherteListe.length === 0) {
        container.innerHTML = `
            <div class="gespeichert-empty">
                <i class="fas fa-book"></i>
                <span>Noch keine Rezepte gespeichert - generiere welche und speichere deine Favoriten.</span>
            </div>`;
        return;
    }

    container.innerHTML = gespeicherteListe.map((r, i) => `
        <div class="gespeichert-item" data-index="${i}" title="Rezept anzeigen">
            <div class="gespeichert-info">
                <div class="gespeichert-name">
                    ${UI.escapeHtml(r.name)}
                    <i class="fas fa-chevron-right gespeichert-chevron"></i>
                </div>
                <div class="gespeichert-meta">
                    <span class="badge badge-info"><i class="fas fa-utensils"></i> ${UI.escapeHtml(r.essens_kategorie)}</span>
                    <span class="badge badge-success"><i class="fas fa-leaf"></i> ${UI.escapeHtml(r.lebensmittel_kategorie)}</span>
                </div>
                ${(r.zutaten || []).length > 0 ? `
                    <div class="gespeichert-zutaten">
                        ${(r.zutaten || []).slice(0, 6).map(z => `<span class="rezept-meta-chip">${UI.escapeHtml(z.name)}</span>`).join('')}
                        ${(r.zutaten || []).length > 6 ? `<span class="rezept-meta-chip">+${(r.zutaten || []).length - 6} weitere</span>` : ''}
                    </div>` : ''}
            </div>
            <div class="gespeichert-actions">
                <button type="button" class="btn-icon btn-danger gespeichert-delete-btn" title="Rezept löschen"
                        data-id="${r.id}">
                    <i class="fas fa-trash"></i>
                </button>
                <div class="gespeichert-datum">${UI.formatDate(r.created_at)}</div>
            </div>
        </div>
    `).join('');

    // Klick auf den Item-Körper öffnet das Rezept
    container.querySelectorAll('.gespeichert-item').forEach(item => {
        item.addEventListener('click', (e) => {
            // Nicht reagieren, wenn der Delete-Button angeklickt wurde
            if (e.target.closest('.gespeichert-delete-btn')) {
                return;
            }
            openGespeichert(parseInt(item.dataset.index, 10));
        });
    });

    // Klick auf den Delete-Button löscht das Rezept
    container.querySelectorAll('.gespeichert-delete-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const id = parseInt(btn.dataset.id, 10);
            deleteRezept(id);
        });
    });
}

// Öffnet ein gespeichertes Rezept als Modal im selben Format wie die KI-Ausgabe
function openGespeichert(index) {
    const r = gespeicherteListe[index];
    if (!r) return;

    const modalEl = document.getElementById('modal');
    modalEl.classList.add('modal-breit');

    Modal.open({
        title: 'Gespeichertes Rezept',
        content: rezeptCard(r, index, {
            source: 'saved',
            saveButton: false,
            gespeichertAm: r.created_at
        }),
        footer: `
            <button type="button" class="btn btn-danger" id="rezeptDeleteBtn" data-id="${r.id}">
                <i class="fas fa-trash"></i> Rezept löschen
            </button>
            <button type="button" class="btn btn-outline" id="gespeichertCloseBtn">
                <i class="fas fa-xmark"></i> Schließen
            </button>
        `,
        onClose: () => modalEl.classList.remove('modal-breit')
    });

    document.getElementById('gespeichertCloseBtn').addEventListener('click', () => Modal.close());
    document.getElementById('rezeptDeleteBtn').addEventListener('click', () => {
        const id = parseInt(document.getElementById('rezeptDeleteBtn').dataset.id, 10);
        deleteRezept(id);
    });
}
