/**
 * KitchenOps - Einkaufsliste-Seite
 * Alle Operationen (Anzeigen, Hinzufügen, Abhaken, Löschen, Leeren)
 * laufen über die Backend-API und sind in der SQLite-DB persistiert.
 */

let artikelList = []; // Einkaufslisten-Artikel aus dem Backend

function PageInit() {
    bindEvents();
    reloadAll();
}

async function reloadAll() {
    Loading.show();
    try {
        await Promise.all([
            loadEinkaufsliste(),
            DataManager.loadLebensmittel(true)
        ]);
    } finally {
        Loading.hide();
    }
    render();
}

async function loadEinkaufsliste() {
    try {
        const data = await api.getAllEinkaufsartikel();
        const list = Array.isArray(data?.artikel) ? data.artikel : [];
        artikelList = list.map(a => ({
            id: a.id ?? 0,
            name: a.name ?? '',
            anzahl: Number(a.anzahl) || 0,
            einheit: a.einheit ?? '',
            notiz: a.notiz ?? '',
            erledigt: !!a.erledigt,
            gefrierfach: !!a.gefrierfach,
            inventar_hinweis: a.inventar_hinweis ?? ''
        }));
    } catch (error) {
        Toast.error(`Fehler beim Laden der Einkaufsliste: ${error.message}`);
        artikelList = [];
    }
}

// -------------------------------------------------------------------------
// Events
// -------------------------------------------------------------------------
function bindEvents() {
    document.getElementById('addItemBtn').addEventListener('click', openAddItemModal);
    document.getElementById('einkaufsTextErfassenBtn').addEventListener('click', erfasseEinkaufsartikelPerText);

    document.getElementById('markAllBoughtBtn').addEventListener('click', markAllBought);

    document.getElementById('clearListBtn').addEventListener('click', async () => {
        if (artikelList.length === 0) {
            Toast.info('Die Liste ist bereits leer');
            return;
        }
        if (!confirm('Möchten Sie die komplette Einkaufsliste löschen?')) return;

        Loading.show();
        try {
            const result = await api.clearEinkaufsliste();
            Toast.success(result?.message || 'Einkaufsliste geleert');
            await loadEinkaufsliste();
        } catch (error) {
            Toast.error(`Fehler beim Leeren: ${error.message}`);
        } finally {
            Loading.hide();
            render();
        }
    });

    // Interaktionen in der Liste (Event-Delegation)
    document.getElementById('einkaufsliste').addEventListener('click', async (e) => {
        const check = e.target.closest('.einkaufsliste-checkbox');
        const boughtBtn = e.target.closest('.item-edit');
        const removeBtn = e.target.closest('.item-remove');
        const clearBoughtBtn = e.target.closest('#clearBoughtBtn');

        if (check) {
            await toggleErledigt(parseInt(check.dataset.id, 10));
        } else if (boughtBtn) {
            await toggleErledigt(parseInt(boughtBtn.dataset.id, 10));
        } else if (removeBtn) {
            await removeArtikel(parseInt(removeBtn.dataset.id, 10));
        } else if (clearBoughtBtn) {
            await clearErledigte();
        }
    });

    // Nachkauf-Vorschläge übernehmen
    document.getElementById('suggestionList').addEventListener('click', async (e) => {
        const addBtn = e.target.closest('.suggestion-add');
        if (!addBtn || addBtn.disabled) return;

        const { name, einheit } = addBtn.dataset;
        const anzahl = addBtn.dataset.anzahl;
        await addArtikel({ name, anzahl: anzahl > 0 ? anzahl : 1, einheit });
    });
}

// -------------------------------------------------------------------------
// API-Operationen
// -------------------------------------------------------------------------
async function addArtikel({ name, anzahl, einheit, notiz = '', gefrierfach = false }) {
    Loading.show();
    try {
        const created = await api.createEinkaufsartikel({
            name: name.trim(),
            anzahl: parseInt(anzahl, 10) || 1,
            einheit,
            notiz,
            gefrierfach
        });
        Toast.success(`"${created?.name || name}" auf der Einkaufsliste`);
        await loadEinkaufsliste();
    } catch (error) {
        Toast.error(`Fehler beim Hinzufügen: ${error.message}`);
    } finally {
        Loading.hide();
        render();
    }
}

async function toggleErledigt(id) {
    const item = artikelList.find(a => a.id === id);
    if (!item) return;

    Loading.show();
    try {
        const response = await api.updateEinkaufsartikel(id, {
            name: item.name,
            anzahl: item.anzahl,
            einheit: item.einheit,
            notiz: item.notiz,
            erledigt: !item.erledigt,
            gefrierfach: item.gefrierfach
        });

        // Backend-Meldung zur automatischen Inventar-Übernahme anzeigen
        if (response?.inventar_hinweis) {
            Toast.success(response.inventar_hinweis);
        }

        await Promise.all([
            loadEinkaufsliste(),
            DataManager.loadLebensmittel(true) // Inventar-Cache aktualisieren
        ]);
    } catch (error) {
        Toast.error(`Fehler beim Abhaken: ${error.message}`);
    } finally {
        Loading.hide();
        render();
    }
}

async function removeArtikel(id) {
    const item = artikelList.find(a => a.id === id);
    if (!item) return;

    Loading.show();
    try {
        await api.deleteEinkaufsartikel(id);
        Toast.info(`"${item.name}" entfernt`);
        await loadEinkaufsliste();
    } catch (error) {
        Toast.error(`Fehler beim Entfernen: ${error.message}`);
    } finally {
        Loading.hide();
        render();
    }
}

async function markAllBought() {
    const open = artikelList.filter(a => !a.erledigt);
    if (open.length === 0) {
        Toast.info('Keine offenen Artikel vorhanden');
        return;
    }

    Loading.show();
    try {
        for (const item of open) {
            await api.updateEinkaufsartikel(item.id, {
                name: item.name,
                anzahl: item.anzahl,
                einheit: item.einheit,
                notiz: item.notiz,
                erledigt: true,
                gefrierfach: item.gefrierfach
            });
        }
        Toast.success(`${open.length} Artikel abgehakt und ins Inventar übernommen`);
        await Promise.all([
            loadEinkaufsliste(),
            DataManager.loadLebensmittel(true)
        ]);
    } catch (error) {
        Toast.error(`Fehler beim Abhaken: ${error.message}`);
        await loadEinkaufsliste();
    } finally {
        Loading.hide();
        render();
    }
}

async function clearErledigte() {
    const bought = artikelList.filter(a => a.erledigt);
    if (bought.length === 0) return;

    Loading.show();
    try {
        for (const item of bought) {
            await api.deleteEinkaufsartikel(item.id);
        }
        Toast.info(`${bought.length} erledigte Artikel entfernt`);
        await loadEinkaufsliste();
    } catch (error) {
        Toast.error(`Fehler beim Entfernen: ${error.message}`);
        await loadEinkaufsliste();
    } finally {
        Loading.hide();
        render();
    }
}

// -------------------------------------------------------------------------
// Rendering
// -------------------------------------------------------------------------
function render() {
    const open = artikelList.filter(a => !a.erledigt);
    const bought = artikelList.filter(a => a.erledigt);

    renderStats(open.length, bought.length);
    renderList(open, bought);
    renderSuggestions(open);
}

function renderStats(openCount, boughtCount) {
    document.getElementById('einkaufStats').innerHTML = `
        <div class="stat-card">
            <div class="stat-card-icon primary"><i class="fas fa-cart-shopping"></i></div>
            <div class="stat-card-value">${openCount}</div>
            <div class="stat-card-label">Offene Artikel</div>
        </div>
        <div class="stat-card">
            <div class="stat-card-icon success"><i class="fas fa-check-double"></i></div>
            <div class="stat-card-value">${boughtCount}</div>
            <div class="stat-card-label">Erledigt</div>
        </div>
    `;
}

function renderList(open, bought) {
    const listEl = document.getElementById('einkaufsliste');

    if (open.length === 0 && bought.length === 0) {
        listEl.innerHTML = UI.emptyState(
            'fa-cart-shopping',
            'Einkaufsliste ist leer',
            'Fügen Sie Artikel hinzu oder übernehmen Sie Vorschläge aus dem kritischen Bestand.'
        );
        return;
    }

    const renderItem = (item, isBought) => `
        <li class="einkaufsliste-item ${isBought ? 'bought' : ''}" data-id="${item.id}">
            <div class="einkaufsliste-checkbox ${isBought ? 'checked' : ''}" data-id="${item.id}" title="Abhaken">
                <i class="fas fa-check"></i>
            </div>
            <div class="einkaufsliste-info">
                <div class="einkaufsliste-name">${UI.escapeHtml(item.name)}${item.gefrierfach ? ' <i class="fas fa-snowflake gefrierfach-row" title="Gefrierfach"></i>' : ''}</div>
                <div class="einkaufsliste-meta">
                    ${item.anzahl > 0 ? `${item.anzahl} ${UI.escapeHtml(item.einheit)}` : UI.escapeHtml(item.einheit)}
                    ${item.notiz ? ` · <em>${UI.escapeHtml(item.notiz)}</em>` : ''}
                </div>
            </div>
            <div class="einkaufsliste-actions">
                ${!isBought ? `
                    <button class="btn-icon item-edit" data-id="${item.id}" title="Als gekauft markieren">
                        <i class="fas fa-circle-check"></i>
                    </button>` : ''}
                <button class="btn-icon item-remove" data-id="${item.id}" title="Entfernen">
                    <i class="fas fa-xmark"></i>
                </button>
            </div>
        </li>`;

    listEl.innerHTML = `
        ${open.map(i => renderItem(i, false)).join('')}
        ${bought.length > 0 ? `
            <li class="einkaufsliste-divider">
                <span>Erledigt (${bought.length})</span>
                <button class="btn btn-outline btn-sm" id="clearBoughtBtn">Erledigte entfernen</button>
            </li>
            ${bought.map(i => renderItem(i, true)).join('')}
        ` : ''}
    `;
}

function renderSuggestions(openItems) {
    const container = document.getElementById('suggestionList');
    const { lebensmittel } = Store.getState();

    const critical = lebensmittel.filter(
        l => l.anzahl <= CONSTANTS.CRITICAL_STOCK_THRESHOLD);

    document.getElementById('suggestionHint').textContent =
        critical.length > 0
            ? `${critical.length} Artikel mit kritischem Bestand`
            : 'Kein kritischer Bestand';

    if (critical.length === 0) {
        container.innerHTML = `
            <div class="suggestion-empty">
                <i class="fas fa-circle-check"></i>
                <span>Alles bestens - kein Nachkauf nötig.</span>
            </div>`;
        return;
    }

    container.innerHTML = critical.map(l => {
        const level = UI.getStockLevel(l.anzahl);
        const onList = openItems.some(
            o => o.name.toLowerCase() === l.name.toLowerCase());

        return `
            <div class="suggestion-item">
                <div class="einkaufsliste-info">
                    <div class="einkaufsliste-name">${UI.escapeHtml(l.name)}</div>
                    <div class="einkaufsliste-meta">
                        Bestand: ${l.anzahl} ${UI.escapeHtml(l.einheit)}
                        · <span class="stock-badge ${level}">${UI.getStockLevelText(level)}</span>
                    </div>
                </div>
                <button class="btn ${onList ? 'btn-outline' : 'btn-primary'} btn-sm suggestion-add"
                        data-name="${UI.escapeHtml(l.name)}"
                        data-anzahl="1"
                        data-einheit="${UI.escapeHtml(l.einheit)}"
                        ${onList ? 'disabled' : ''}>
                    ${onList ? '<i class="fas fa-check"></i> auf Liste' : '<i class="fas fa-plus"></i> Übernehmen'}
                </button>
            </div>`;
    }).join('');
}

// -------------------------------------------------------------------------
// Artikel hinzufügen (Modal)
// -------------------------------------------------------------------------
function openAddItemModal() {
    Modal.open({
        title: 'Artikel zur Einkaufsliste hinzufügen',
        content: `
            <form id="einkaufForm" novalidate>
                <div class="form-group">
                    <label for="ekName">Artikel *</label>
                    <input type="text" id="ekName" class="form-control" placeholder="z.B. Mehl" autocomplete="off" required>
                    <div class="form-error">Bitte geben Sie einen Artikel ein</div>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label for="ekAnzahl">Menge</label>
                        <input type="number" id="ekAnzahl" class="form-control" min="0" value="1">
                    </div>
                    <div class="form-group">
                        <label for="ekEinheit">Einheit</label>
                        <select id="ekEinheit" class="form-control form-select">
                            ${CONSTANTS.COMMON_UNITS.map(u => `<option value="${u}" ${u === 'Stück' ? 'selected' : ''}>${u}</option>`).join('')}
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <label for="ekNote">Notiz (optional)</label>
                    <input type="text" id="ekNote" class="form-control" placeholder="z.B. Biomarkt, große Packung">
                </div>
                <div class="form-group">
                    <label>Lagerung</label>
                    <label class="checkbox-chip ek-gefrierfach">
                        <input type="checkbox" id="ekGefrierfach">
                        <span><i class="fas fa-snowflake"></i> Gefrierfach (Kühltruhe)</span>
                    </label>
                </div>
                <div class="modal-footer no-border">
                    <button type="button" class="btn btn-outline" id="ekCancelBtn">Abbrechen</button>
                    <button type="submit" class="btn btn-primary"><i class="fas fa-plus"></i> Hinzufügen</button>
                </div>
            </form>`
    });

    document.getElementById('ekCancelBtn').addEventListener('click', () => Modal.close());
    document.getElementById('einkaufForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = document.getElementById('ekName').value.trim();
        if (!name) {
            document.getElementById('ekName').closest('.form-group').classList.add('error');
            Toast.warning('Bitte geben Sie einen Artikel ein');
            return;
        }

        const notiz = document.getElementById('ekNote').value.trim();
        const gefrierfach = document.getElementById('ekGefrierfach')?.checked || false;
        const anzahl = document.getElementById('ekAnzahl').value;
        const einheit = document.getElementById('ekEinheit').value;

        Modal.close();
        await addArtikel({ name, anzahl, einheit, notiz, gefrierfach });
    });

    setTimeout(() => document.getElementById('ekName')?.focus(), 50);
}

// -------------------------------------------------------------------------
// KI-Text-Erfassung für Einkaufsartikel
// -------------------------------------------------------------------------
async function erfasseEinkaufsartikelPerText() {
    const textInput = document.getElementById('einkaufsTextInput');
    const text = textInput.value.trim();

    if (!text) {
        Toast.warning('Bitte geben Sie einen Text ein');
        return;
    }

    const btn = document.getElementById('einkaufsTextErfassenBtn');
    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Verarbeite...';
    textInput.disabled = true;

    try {
        const data = await api.erfasseEinkaufsartikel(text);
        const artikel = Array.isArray(data?.artikel) ? data.artikel : [];
        
        Toast.success(`${artikel.length} Artikel zur Einkaufsliste hinzugefügt`);
        textInput.value = '';
        
        // Liste neu laden
        await loadEinkaufsliste();
        render();
    } catch (error) {
        Toast.error(`Fehler bei der Text-Erfassung: ${error.message}`);
    } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-wand-magic-sparkles"></i> Artikel aus Text erfassen';
        textInput.disabled = false;
        textInput.focus();
    }
}
