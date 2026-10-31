/**
 * KitchenOps - Dashboard-Seite
 * Zeigt Statistiken, Schnellaktionen und die neuesten Lebensmittel.
 */

function PageInit() {
    render();
    DataManager.loadLebensmittel().then(render);

    Store.subscribe(() => render());
}

function render() {
    const { lebensmittel, serverConnected } = Store.getState();

    // Verbindungsbanner
    document.getElementById('connectionBanner')
        .classList.toggle('hidden', serverConnected !== false);

    renderStats(lebensmittel);
    renderQuickActions();
    renderRecent(lebensmittel);
}

function renderStats(lebensmittel) {
    const total = lebensmittel.length;
    const lowStock = lebensmittel.filter(l => l.anzahl <= CONSTANTS.LOW_STOCK_THRESHOLD).length;
    const critical = lebensmittel.filter(l => l.anzahl <= CONSTANTS.CRITICAL_STOCK_THRESHOLD).length;
    const purchase = lebensmittel.filter(l => l.needsPurchase).length;
    const totalAmount = lebensmittel.reduce((sum, l) => sum + l.anzahl, 0);

    document.getElementById('statsGrid').innerHTML = `
        <div class="stat-card">
            <div class="stat-card-icon primary"><i class="fas fa-box-open"></i></div>
            <div class="stat-card-value">${total}</div>
            <div class="stat-card-label">Lebensmittel</div>
        </div>
        <div class="stat-card">
            <div class="stat-card-icon warning"><i class="fas fa-exclamation-triangle"></i></div>
            <div class="stat-card-value">${lowStock}</div>
            <div class="stat-card-label">Niedriger Bestand</div>
        </div>
        <div class="stat-card">
            <div class="stat-card-icon error"><i class="fas fa-triangle-exclamation"></i></div>
            <div class="stat-card-value">${critical}</div>
            <div class="stat-card-label">Kritischer Bestand</div>
        </div>
        <div class="stat-card">
            <div class="stat-card-icon secondary"><i class="fas fa-cart-shopping"></i></div>
            <div class="stat-card-value">${purchase}</div>
            <div class="stat-card-label">Einkaufsliste</div>
        </div>
        <div class="stat-card">
            <div class="stat-card-icon success"><i class="fas fa-chart-line"></i></div>
            <div class="stat-card-value">${totalAmount.toLocaleString('de-DE')}</div>
            <div class="stat-card-label">Gesamtmenge</div>
        </div>
    `;
}

function renderQuickActions() {
    document.getElementById('quickActions').innerHTML = `
        <a href="lebensmittel.html" class="quick-action-card">
            <i class="fas fa-plus quick-action-icon primary"></i>
            <span class="quick-action-title">Hinzufügen</span>
            <span class="quick-action-desc">Neues Lebensmittel anlegen</span>
        </a>
        <a href="einkaufsliste.html" class="quick-action-card">
            <i class="fas fa-cart-shopping quick-action-icon secondary"></i>
            <span class="quick-action-title">Einkaufsliste</span>
            <span class="quick-action-desc">Nachkaufplanung anzeigen</span>
        </a>
        <a href="inventar.html" class="quick-action-card">
            <i class="fas fa-clipboard-list quick-action-icon success"></i>
            <span class="quick-action-title">Inventar</span>
            <span class="quick-action-desc">Bestände nach Kategorie</span>
        </a>
        <a href="einstellungen.html" class="quick-action-card">
            <i class="fas fa-gear quick-action-icon warning"></i>
            <span class="quick-action-title">Einstellungen</span>
            <span class="quick-action-desc">Server & Design verwalten</span>
        </a>
    `;
}

function renderRecent(lebensmittel) {
    const container = document.getElementById('recentLebensmittel');

    if (lebensmittel.length === 0) {
        container.innerHTML = UI.emptyState(
            'fa-box-open',
            'Noch keine Lebensmittel',
            'Fügen Sie Ihr erstes Lebensmittel hinzu, um zu starten.',
            `<a href="lebensmittel.html" class="btn btn-primary mt-md"><i class="fas fa-plus"></i> Lebensmittel hinzufügen</a>`
        );
        return;
    }

    // Neueste zuerst (nach updatedAt/createdAt, Fallback: ID)
    const sorted = [...lebensmittel].sort((a, b) => {
        const da = new Date(b.updatedAt || b.createdAt || 0).getTime();
        const db = new Date(a.updatedAt || a.createdAt || 0).getTime();
        return da - db || b.id - a.id;
    });

    container.innerHTML = sorted.slice(0, 6).map(l => UI.lebensmittelCard(l, false)).join('');

    // Karten verlinken zur Lebensmittel-Detailansicht
    container.querySelectorAll('.lebensmittel-card').forEach(card => {
        card.addEventListener('click', () => {
            location.href = `lebensmittel.html?id=${card.dataset.id}`;
        });
    });
}
