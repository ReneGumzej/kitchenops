/**
 * KitchenOps - Lebensmittel-Seite
 * Funktionen: Suche, Filter, Grid-/Tabellenansicht, CRUD, Detailansicht via ?id=
 */

let currentView = localStorage.getItem('lmView') || 'grid';

function PageInit() {
    // Query-Parameter: ?q=Suchbegriff und ?id=Detail
    const params = new URLSearchParams(location.search);
    const query = params.get('q');
    const detailId = params.get('id');

    if (query) {
        Store.setState({ searchQuery: query });
        document.getElementById('lmSearch').value = query;
    }

    bindEvents();
    applyView(currentView, false);

    DataManager.loadLebensmittel().then(() => {
        render();
        if (detailId) openDetail(parseInt(detailId, 10));
    });

    Store.subscribe(() => render());
}

// -------------------------------------------------------------------------
// Events
// -------------------------------------------------------------------------
function bindEvents() {
    // Suche (mit kleinem Debounce)
    let debounce;
    document.getElementById('lmSearch').addEventListener('input', (e) => {
        clearTimeout(debounce);
        debounce = setTimeout(() => {
            Store.setState({ searchQuery: e.target.value });
        }, 200);
    });

    // Filter-Chips
    document.querySelectorAll('.chip').forEach(chip => {
        chip.addEventListener('click', () => {
            document.querySelectorAll('.chip').forEach(c => c.classList.remove('active'));
            chip.classList.add('active');
            Store.setState({ filter: chip.dataset.filter });
        });
    });

    // Ansichtswechsel Grid/Tabelle
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', () => applyView(btn.dataset.view, true));
    });

    // Neu anlegen
    document.getElementById('addLebensmittelBtn').addEventListener('click', () => LebensmittelForm.open(null));

    // Auto-Kategorisieren (Regeln + KI-Batch im Backend)
    document.getElementById('kategorisierenBtn').addEventListener('click', autoKategorisieren);

    // KI-Schnellerfassung per Freitext
    document.getElementById('kiErfassungBtn').addEventListener('click', kiErfassen);
    document.getElementById('kiErfassungText').addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            kiErfassen();
        }
    });

    // Bearbeiten / Löschen (Event-Delegation, gilt für Grid und Tabelle)
    const content = document.querySelector('.page-content');
    content.addEventListener('click', (e) => {
        const editBtn = e.target.closest('.edit-btn');
        const deleteBtn = e.target.closest('.delete-btn');
        const row = e.target.closest('tr[data-id]');

        if (editBtn) {
            e.stopPropagation();
            LebensmittelForm.open(parseInt(editBtn.dataset.id, 10));
        } else if (deleteBtn) {
            e.stopPropagation();
            confirmDelete(parseInt(deleteBtn.dataset.id, 10));
        } else if (row) {
            openDetail(parseInt(row.dataset.id, 10));
        } else {
            const card = e.target.closest('.lebensmittel-card');
            if (card && !e.target.closest('.lebensmittel-actions')) {
                openDetail(parseInt(card.dataset.id, 10));
            }
        }
    });
}

// -------------------------------------------------------------------------
// Ansicht
// -------------------------------------------------------------------------
function applyView(view, save = true) {
    currentView = view;
    if (save) localStorage.setItem('lmView', view);

    document.querySelectorAll('.view-btn').forEach(b =>
        b.classList.toggle('active', b.dataset.view === view));

    document.getElementById('lmGrid').classList.toggle('hidden', view !== 'grid');
    document.getElementById('lmTableWrap').classList.toggle('hidden', view !== 'table');
}

// -------------------------------------------------------------------------
// Rendering
// -------------------------------------------------------------------------
function render() {
    const { lebensmittel, searchQuery, filter } = Store.getState();
    const filtered = applyFilters(lebensmittel, searchQuery, filter);

    document.getElementById('resultCount').textContent =
        `${filtered.length} von ${lebensmittel.length} Einträgen`;

    // Kartenansicht
    const grid = document.getElementById('lmGrid');
    if (filtered.length === 0) {
        grid.innerHTML = UI.emptyState(
            'fa-magnifying-glass',
            'Keine Lebensmittel gefunden',
            lebensmittel.length === 0
                ? 'Legen Sie Ihr erstes Lebensmittel an.'
                : 'Versuchen Sie einen anderen Suchbegriff oder Filter.'
        );
    } else {
        grid.innerHTML = filtered.map(l => UI.lebensmittelCard(l, true)).join('');
    }

    // Tabellenansicht
    const tbody = document.getElementById('lmTableBody');
    if (filtered.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="table-empty">Keine Lebensmittel gefunden</td></tr>`;
    } else {
        tbody.innerHTML = filtered.map(l => UI.lebensmittelRow(l)).join('');
    }
}

// Auto-Kategorisierung: Backend ordnet per Regeln zu, Rest per KI-Batch
async function autoKategorisieren() {
    const btn = document.getElementById('kategorisierenBtn');
    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Kategorisiere...';

    try {
        const result = await api.kategorisierenLebensmittel();
        const teile = [];
        if (result.per_regeln > 0) teile.push(`${result.per_regeln} per Regeln`);
        if (result.per_llm > 0) teile.push(`${result.per_llm} per KI`);
        if (result.offen > 0) teile.push(`${result.offen} offen`);
        Toast.success(`Kategorisierung: ${teile.join(', ') || 'alles bereits zugeordnet'}`);
        await DataManager.loadLebensmittel(true);
    } catch (error) {
        Toast.error(`Kategorisierung fehlgeschlagen: ${error.message}`);
    } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-wand-magic-sparkles"></i> Auto-Kategorisieren';
    }
}

// KI-Schnellerfassung: Freitext -> strukturierte Lebensmittel in der DB
async function kiErfassen() {
    const textEl = document.getElementById('kiErfassungText');
    const btn = document.getElementById('kiErfassungBtn');
    const text = textEl.value.trim();

    if (!text) {
        Toast.warning('Bitte beschreiben Sie, was hinzugefügt werden soll');
        textEl.focus();
        return;
    }

    btn.disabled = true;
    btn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Erfasse...';

    try {
        const result = await api.erfasseLebensmittel(text);
        const erfasst = result?.lebensmittel || [];

        if (erfasst.length === 0) {
            Toast.warning('Keine Lebensmittel im Text erkannt');
        } else {
            const aufzaehlung = erfasst
                .map(l => `${l.name} (${l.anzahl} ${l.einheit}${l.gefrierfach ? ', Gefrierfach' : ''})`)
                .join(', ');
            Toast.success(`Erfasst: ${aufzaehlung}`, 5000);
            textEl.value = '';
            await DataManager.loadLebensmittel(true);
        }
    } catch (error) {
        Toast.error(`Erfassung fehlgeschlagen: ${error.message}`);
    } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="fas fa-wand-magic-sparkles"></i> Erfassen';
    }
}

function applyFilters(list, query, filter) {
    let result = [...list];

    if (query && query.trim()) {
        const q = query.trim().toLowerCase();
        result = result.filter(l =>
            l.name.toLowerCase().includes(q) ||
            l.einheit.toLowerCase().includes(q));
    }

    switch (filter) {
        case 'low':
            result = result.filter(l =>
                l.anzahl > CONSTANTS.CRITICAL_STOCK_THRESHOLD &&
                l.anzahl <= CONSTANTS.LOW_STOCK_THRESHOLD);
            break;
        case 'critical':
            result = result.filter(l => l.anzahl <= CONSTANTS.CRITICAL_STOCK_THRESHOLD);
            break;
        case 'purchase':
            result = result.filter(l => l.needsPurchase || l.anzahl <= 0);
            break;
    }

    // Alphabetisch sortieren
    result.sort((a, b) => a.name.localeCompare(b.name, 'de'));
    return result;
}

// -------------------------------------------------------------------------
// Detailansicht (Modal)
// -------------------------------------------------------------------------
function openDetail(id) {
    const item = Store.getState().lebensmittel.find(l => l.id === id);
    if (!item) {
        Toast.warning('Lebensmittel nicht gefunden');
        return;
    }

    const level = UI.getStockLevel(item.anzahl);
    const progress = Math.min((item.anzahl / 10) * 100, 100);

    Modal.open({
        title: item.name,
        content: `
            <div class="lebensmittel-detail">
                <div class="detail-icon"><i class="fas fa-basket-shopping"></i></div>

                <div class="detail-hero">
                    <span class="detail-amount">${item.anzahl}</span>
                    <span class="detail-unit">${UI.escapeHtml(item.einheit)}</span>
                </div>

                <div class="lebensmittel-progress detail-progress">
                    <div class="lebensmittel-progress-bar ${level}" style="width: ${progress}%"></div>
                </div>
                <span class="stock-badge ${level}">${UI.getStockLevelText(level)}</span>

                <div class="detail-rows">
                    <div class="detail-row">
                        <span class="detail-label">ID</span>
                        <span class="detail-value">#${item.id}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Lagerung</span>
                        <span class="detail-value">${item.gefrierfach
                            ? '<span class="badge badge-gefrierfach"><i class="fas fa-snowflake"></i> Gefrierfach (Kühltruhe)</span>'
                            : 'Normal (Vorratsschrank)'}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Küchenrichtung</span>
                        <span class="detail-value">${UI.kategorieBadge(item.lebensmittel_kategorie) || '<span class="text-muted">nicht zugeordnet</span>'}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Mahlzeiten</span>
                        <span class="detail-value">${(item.mahlzeiten || []).length > 0 ? item.mahlzeiten.join(', ') : '<span class="text-muted">nicht zugeordnet</span>'}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Erstellt</span>
                        <span class="detail-value">${UI.formatDate(item.createdAt)}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Aktualisiert</span>
                        <span class="detail-value">${UI.formatDate(item.updatedAt)}</span>
                    </div>
                </div>
            </div>
        `,
        footer: `
            <button class="btn btn-outline" id="detailEditBtn"><i class="fas fa-edit"></i> Bearbeiten</button>
            <button class="btn btn-error" id="detailDeleteBtn"><i class="fas fa-trash"></i> Löschen</button>
        `
    });

    document.getElementById('detailEditBtn').addEventListener('click', () => LebensmittelForm.open(id));
    document.getElementById('detailDeleteBtn').addEventListener('click', () => confirmDelete(id));
}
