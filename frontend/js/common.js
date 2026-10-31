/**
 * KitchenOps - Gemeinsame Logik für alle Seiten
 * Enthält: Konstanten, Store, Toast, Modal, Loading, Theme, Layout, DataManager
 * Wird auf jeder HTML-Seite nach api.js eingebunden.
 */

// =========================================================================
// KONSTANTEN
// =========================================================================
const CONSTANTS = {
    // Aus den Einstellungen (localStorage), sonst Standardwerte
    LOW_STOCK_THRESHOLD: (() => {
        const v = parseInt(localStorage.getItem('lowStockThreshold'), 10);
        return isNaN(v) ? 5 : v;
    })(),
    CRITICAL_STOCK_THRESHOLD: (() => {
        const v = parseInt(localStorage.getItem('criticalStockThreshold'), 10);
        return isNaN(v) ? 2 : v;
    })(),

    COMMON_UNITS: ['g', 'kg', 'mg', 'l', 'ml', 'Stück', 'Dose', 'Glas', 'Flasche', 'Packung', 'Beutel'],

    // Kategorien für die Zuordnung der Lebensmittel
    KUECHENRICHTUNGEN: ['Pasta', 'Rindfleisch', 'Hühnerfleisch', 'Schweinefleisch', 'Vegetarisch', 'Vegan', 'Salat'],
    MAHLZEITEN: ['Frühstück', 'Mittagessen', 'Abendessen', 'Snack'],

    ERROR_MESSAGES: {
        deleteConfirm: 'Möchten Sie dieses Lebensmittel wirklich löschen?',
        network: 'Netzwerkfehler. Läuft der Backend-Server?'
    }
};

// =========================================================================
// STORE - geteilter Zustand pro Seite (Lebensmittel-Cache)
// =========================================================================
const Store = {
    state: {
        lebensmittel: [],
        loading: false,
        error: null,
        searchQuery: '',
        filter: 'all',
        theme: localStorage.getItem('theme') || 'light',
        serverConnected: null
    },

    subscribers: [],

    getState() {
        return { ...this.state };
    },

    setState(newState) {
        this.state = { ...this.state, ...newState };
        this.subscribers.forEach(cb => cb(this.getState()));
    },

    subscribe(callback) {
        this.subscribers.push(callback);
    }
};

// =========================================================================
// TOAST NOTIFICATIONS
// =========================================================================
const Toast = {
    container: null,

    init() {
        this.container = document.getElementById('toastContainer');
    },

    show(message, type = 'info', duration = 3000) {
        if (!this.container) return;

        const iconMap = {
            success: 'fa-check-circle',
            error: 'fa-exclamation-circle',
            warning: 'fa-exclamation-triangle',
            info: 'fa-info-circle'
        };

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
            <i class="fas ${iconMap[type] || 'fa-info-circle'} toast-icon"></i>
            <span>${message}</span>
        `;

        this.container.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'toastSlideIn 0.25s reverse';
            setTimeout(() => toast.remove(), 250);
        }, duration);
    },

    success(m, d = 3000) { this.show(m, 'success', d); },
    error(m, d = 4000) { this.show(m, 'error', d); },
    warning(m, d = 4000) { this.show(m, 'warning', d); },
    info(m, d = 3000) { this.show(m, 'info', d); }
};

// =========================================================================
// MODAL
// =========================================================================
const Modal = {
    overlay: null,
    titleEl: null,
    bodyEl: null,
    footerEl: null,
    closeCallback: null,

    init() {
        this.overlay = document.getElementById('modalOverlay');
        this.titleEl = document.getElementById('modalTitle');
        this.bodyEl = document.getElementById('modalBody');
        this.footerEl = document.getElementById('modalFooter');

        document.getElementById('modalClose').addEventListener('click', () => this.close());
        // Absichtlich KEIN Schließen bei Klick auf das Overlay (außerhalb des
        // Modals): das Modal bleibt bewusst offen und schließt nur über
        // den X-Button, explizite Schließen-/Abbrechen-Buttons oder Escape.
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') this.close();
        });
    },

    open({ title = '', content = '', footer = '', onClose = null }) {
        this.titleEl.textContent = title;
        this.bodyEl.innerHTML = content;
        this.footerEl.innerHTML = footer;
        this.closeCallback = onClose;
        this.overlay.classList.remove('hidden');
    },

    close() {
        this.overlay.classList.add('hidden');
        this.bodyEl.innerHTML = '';
        this.footerEl.innerHTML = '';
        if (this.closeCallback) {
            this.closeCallback();
            this.closeCallback = null;
        }
    }
};

// =========================================================================
// LOADING OVERLAY
// =========================================================================
const Loading = {
    overlay: null,

    init() {
        this.overlay = document.getElementById('loadingOverlay');
    },

    show() { this.overlay?.classList.remove('hidden'); },
    hide() { this.overlay?.classList.add('hidden'); }
};

// =========================================================================
// THEME
// =========================================================================
const Theme = {
    init() {
        this.apply(localStorage.getItem('theme') || 'light');
        document.getElementById('themeToggle')?.addEventListener('click', () => {
            const current = document.documentElement.getAttribute('data-theme');
            this.set(current === 'dark' ? 'light' : 'dark');
        });
    },

    set(theme) {
        localStorage.setItem('theme', theme);
        this.apply(theme);
    },

    apply(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        const icon = document.querySelector('#themeToggle i');
        if (icon) icon.className = theme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
    }
};

// =========================================================================
// LAYOUT (Sidebar, Mobile-Navigation, aktive Nav-Links, Suche)
// =========================================================================
const Layout = {
    init() {
        // Sidebar-Toggle (Desktop + Mobile)
        const sidebar = document.getElementById('sidebar');
        const mainContent = document.querySelector('.main-content');

        document.getElementById('sidebarToggle')?.addEventListener('click', () => {
            sidebar.classList.toggle('open');
            mainContent?.classList.toggle('sidebar-open');
        });
        document.getElementById('mobileNavToggle')?.addEventListener('click', () => {
            sidebar.classList.add('open');
        });
        document.getElementById('sidebarClose')?.addEventListener('click', () => {
            sidebar.classList.remove('open');
            mainContent?.classList.remove('sidebar-open');
        });

        // Aktiven Nav-Link anhand des Dateinamens markieren
        const currentPage = (location.pathname.split('/').pop() || 'index.html').toLowerCase();
        document.querySelectorAll('.nav-link').forEach(link => {
            const target = (link.getAttribute('href') || '').toLowerCase();
            const matches =
                (target === 'index.html' && (currentPage === '' || currentPage === 'index.html')) ||
                target === currentPage;
            if (matches) link.classList.add('active');
        });

        // Globale Suche: leitet zur Lebensmittel-Seite mit Query weiter
        const searchInput = document.getElementById('globalSearch');
        searchInput?.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && e.target.value.trim()) {
                location.href = `lebensmittel.html?q=${encodeURIComponent(e.target.value.trim())}`;
            }
        });
    }
};

// =========================================================================
// HILFSFUNKTIONEN (UI-Bausteine)
// =========================================================================
const UI = {
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text ?? '';
        return div.innerHTML;
    },

    getStockLevel(anzahl) {
        if (anzahl <= CONSTANTS.CRITICAL_STOCK_THRESHOLD) return 'stock-critical';
        if (anzahl <= CONSTANTS.LOW_STOCK_THRESHOLD) return 'stock-low';
        return 'stock-normal';
    },

    getStockLevelClass(level) {
        return { 'stock-normal': 'normal', 'stock-low': 'low', 'stock-critical': 'critical' }[level] || 'normal';
    },

    getStockLevelText(level) {
        const map = { 'stock-normal': 'Normal', 'stock-low': 'Niedrig', 'stock-critical': 'Kritisch' };
        return map[level] || 'Unbekannt';
    },

    formatDate(value) {
        if (!value) return '-';
        try {
            return new Date(value).toLocaleString('de-DE', {
                dateStyle: 'medium',
                timeStyle: 'short'
            });
        } catch {
            return '-';
        }
    },

    // Badge für eine Küchenrichtung (Farbe je Kategorie)
    kategorieBadge(kategorie) {
        if (!kategorie) return '';
        const klassen = {
            'Pasta': 'badge-primary',
            'Rindfleisch': 'badge-error',
            'Hühnerfleisch': 'badge-warning',
            'Schweinefleisch': 'badge-error',
            'Vegetarisch': 'badge-success',
            'Vegan': 'badge-success',
            'Salat': 'badge-info'
        };
        const klasse = klassen[kategorie] || 'badge-info';
        const icons = {
            'Pasta': 'fa-plate-wheat',
            'Rindfleisch': 'fa-drumstick-bite',
            'Hühnerfleisch': 'fa-drumstick',
            'Schweinefleisch': 'fa-bacon',
            'Vegetarisch': 'fa-leaf',
            'Vegan': 'fa-seedling',
            'Salat': 'fa-leaf'
        };
        return `<span class="badge ${klasse} kategorie-badge"><i class="fas ${icons[kategorie] || 'fa-tag'}"></i> ${this.escapeHtml(kategorie)}</span>`;
    },

    // Lebensmittel-Karte (Grid-Ansicht)
    lebensmittelCard(l, withActions = true) {
        const level = this.getStockLevel(l.anzahl);
        const progress = Math.min((l.anzahl / 10) * 100, 100);

        const actions = withActions ? `
            <div class="lebensmittel-actions">
                <button class="lebensmittel-action-btn edit-btn" data-id="${l.id}" title="Bearbeiten">
                    <i class="fas fa-edit"></i> Bearbeiten
                </button>
                <button class="lebensmittel-action-btn delete-btn" data-id="${l.id}" title="Löschen">
                    <i class="fas fa-trash"></i> Löschen
                </button>
            </div>` : '';

        return `
            <div class="lebensmittel-card" data-id="${l.id}">
                <div class="lebensmittel-card-header">
                    <div class="lebensmittel-icon"><i class="fas fa-basket-shopping"></i></div>
                    <div class="lebensmittel-info">
                        <div class="lebensmittel-name">${this.escapeHtml(l.name)}</div>
                        <div class="lebensmittel-meta">
                            <span class="lebensmittel-amount">${l.anzahl} ${this.escapeHtml(l.einheit)}</span>
                        </div>
                    </div>
                    ${l.needsPurchase ? '<span class="badge badge-primary mt-xs"><i class="fas fa-cart-plus"></i> Kaufen</span>' : ''}
                </div>
                <div class="lebensmittel-kategorien">
                    ${l.gefrierfach ? '<span class="badge badge-gefrierfach"><i class="fas fa-snowflake"></i> Gefrierfach</span>' : ''}
                    ${this.kategorieBadge(l.lebensmittel_kategorie)}
                    ${(l.mahlzeiten || []).slice(0, 2).map(m => `<span class="rezept-meta-chip">${this.escapeHtml(m)}</span>`).join('')}
                </div>
                <div class="lebensmittel-card-body">
                    <div class="lebensmittel-progress">
                        <div class="lebensmittel-progress-bar ${level}" style="width: ${progress}%"></div>
                    </div>
                    <span class="stock-badge ${level}">${this.getStockLevelText(level)}</span>
                </div>
                ${actions}
            </div>`;
    },

    // Lebensmittel-Zeile (Tabellen-Ansicht)
    lebensmittelRow(l) {
        const level = this.getStockLevel(l.anzahl);
        return `
            <tr data-id="${l.id}">
                <td class="cell-name" title="Details anzeigen">
                    <i class="fas fa-basket-shopping row-icon"></i>
                    <span>${this.escapeHtml(l.name)}</span>
                    ${l.gefrierfach ? ' <i class="fas fa-snowflake gefrierfach-row" title="Gefrierfach"></i>' : ''}
                </td>
                <td>${l.anzahl}</td>
                <td>${this.escapeHtml(l.einheit)}</td>
                <td>${this.kategorieBadge(l.lebensmittel_kategorie) || '<span class="text-muted text-sm">—</span>'}</td>
                <td><span class="stock-badge ${level}">${this.getStockLevelText(level)}</span></td>
                <td>
                    <div class="table-actions">
                        <button class="btn-icon edit-btn" data-id="${l.id}" title="Bearbeiten">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn-icon delete-btn" data-id="${l.id}" title="Löschen">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>`;
    },

    emptyState(icon, title, text, actionHtml = '') {
        return `
            <div class="empty-state">
                <i class="fas ${icon} empty-state-icon"></i>
                <h3>${title}</h3>
                <p>${text}</p>
                ${actionHtml}
            </div>`;
    }
};

// =========================================================================
// DATA MANAGER - Backend-Kommunikation (geteilt zwischen Seiten)
// =========================================================================
const DataManager = {
    // GET /lebensmittel → normalisiert die Antwort in ein Array
    async loadLebensmittel(silent = false) {
        try {
            if (!silent) Loading.show();
            const data = await api.getAllLebensmittel();

            const list = Array.isArray(data?.lebensmittel)
                ? data.lebensmittel
                : (Array.isArray(data) ? data : []);

            const formatted = list.map(l => ({
                id: l.id ?? 0,
                name: l.name ?? '',
                anzahl: Number(l.anzahl) || 0,
                einheit: l.einheit ?? '',
                lebensmittel_kategorie: l.lebensmittel_kategorie || '',
                mahlzeiten: Array.isArray(l.mahlzeiten) ? l.mahlzeiten : [],
                gefrierfach: !!l.gefrierfach,
                needsPurchase: !!l.needsPurchase,
                createdAt: l.created_at || l.createdAt || null,
                updatedAt: l.updated_at || l.updatedAt || null
            }));

            Store.setState({ lebensmittel: formatted, loading: false, error: null, serverConnected: true });

            // Kein Erfolgs-Toast beim Laden: Seitenaufrufe sollen sich nicht
            // mit Benachrichtigungen bemerkbar machen. Fehler bleiben sichtbar.
            return formatted;
        } catch (error) {
            Store.setState({ loading: false, error: error.message, serverConnected: false });
            if (!silent) Toast.error(`Fehler beim Laden: ${error.message}`);
            return [];
        } finally {
            if (!silent) Loading.hide();
        }
    },

    async createLebensmittel(item) {
        Loading.show();
        try {
            await api.createLebensmittel({
                name: item.name,
                anzahl: parseInt(item.anzahl, 10) || 0,
                einheit: item.einheit,
                lebensmittel_kategorie: item.lebensmittel_kategorie || '',
                mahlzeiten: item.mahlzeiten || [],
                gefrierfach: !!item.gefrierfach
            });
            Toast.success(`"${item.name}" wurde hinzugefügt`);
            await this.loadLebensmittel(true);
        } finally {
            Loading.hide();
        }
    },

    async updateLebensmittel(id, item) {
        Loading.show();
        try {
            await api.updateLebensmittel(id, {
                name: item.name,
                anzahl: parseInt(item.anzahl, 10) || 0,
                einheit: item.einheit,
                lebensmittel_kategorie: item.lebensmittel_kategorie || '',
                mahlzeiten: item.mahlzeiten || [],
                gefrierfach: !!item.gefrierfach
            });
            Toast.success(`"${item.name}" wurde aktualisiert`);
            await this.loadLebensmittel(true);
        } finally {
            Loading.hide();
        }
    },

    async deleteLebensmittel(id) {
        Loading.show();
        try {
            await api.deleteLebensmittel(id);
            Toast.success('Lebensmittel gelöscht');
            await this.loadLebensmittel(true);
        } finally {
            Loading.hide();
        }
    },

    async testConnection(url = null) {
        if (url) api.setBaseUrl(url);
        // ping() gibt true/false zurück und wirft KEINE Fehler -
        // der Rückgabewert muss ausgewertet werden.
        const erreichbar = await api.ping();
        if (erreichbar) {
            Store.setState({ serverConnected: true });
            Toast.success('Verbindung erfolgreich');
            return true;
        }
        Store.setState({ serverConnected: false });
        Toast.error('Verbindung fehlgeschlagen');
        return false;
    }
};

// =========================================================================
// FORMULAR (Lebensmittel anlegen/bearbeiten - von mehreren Seiten genutzt)
// =========================================================================
const LebensmittelForm = {
    render(item = null) {
        const isEdit = !!item;
        const kategorie = item?.lebensmittel_kategorie || '';
        const mahlzeiten = item?.mahlzeiten || [];

        return `
            <form id="lebensmittelForm" novalidate>
                <div class="form-group">
                    <label for="lmName">Name *</label>
                    <input type="text" id="lmName" class="form-control" value="${item ? UI.escapeHtml(item.name) : ''}" autocomplete="off" required>
                    <div class="form-error">Bitte geben Sie einen Namen ein</div>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label for="lmAnzahl">Menge *</label>
                        <input type="number" id="lmAnzahl" class="form-control" value="${item ? item.anzahl : ''}" min="0" required>
                        <div class="form-error">Bitte geben Sie eine gültige Menge ein</div>
                    </div>
                    <div class="form-group">
                        <label for="lmEinheit">Einheit *</label>
                        <select id="lmEinheit" class="form-control form-select" required>
                            ${CONSTANTS.COMMON_UNITS.map(u => `<option value="${u}" ${item && item.einheit === u ? 'selected' : ''}>${u}</option>`).join('')}
                        </select>
                        <div class="form-error">Bitte wählen Sie eine Einheit</div>
                    </div>
                </div>
                <div class="form-group">
                    <label for="lmKategorie">Küchenrichtung</label>
                    <select id="lmKategorie" class="form-control form-select">
                        <option value="">Automatisch (Regeln / KI)</option>
                        ${CONSTANTS.KUECHENRICHTUNGEN.map(k =>
                            `<option value="${k}" ${kategorie === k ? 'selected' : ''}>${k}</option>`).join('')}
                    </select>
                    <p class="text-muted text-sm mt-xs">Ohne Auswahl wird die Kategorie automatisch ermittelt.</p>
                </div>
                <div class="form-group">
                    <label>Mahlzeiten</label>
                    <div class="mahlzeiten-checkboxen">
                        ${CONSTANTS.MAHLZEITEN.map(m => `
                            <label class="checkbox-chip">
                                <input type="checkbox" class="lm-mahlzeit" value="${m}" ${mahlzeiten.includes(m) ? 'checked' : ''}>
                                <span>${m}</span>
                            </label>`).join('')}
                    </div>
                    <p class="text-muted text-sm mt-xs">Ohne Auswahl werden passende Mahlzeiten automatisch ermittelt.</p>
                </div>
                <div class="form-group">
                    <label>Lagerung</label>
                    <label class="checkbox-chip lm-gefrierfach">
                        <input type="checkbox" id="lmGefrierfach" ${item?.gefrierfach ? 'checked' : ''}>
                        <span><i class="fas fa-snowflake"></i> Gefrierfach (Kühltruhe)</span>
                    </label>
                </div>
                <div class="modal-footer no-border">
                    <button type="button" class="btn btn-outline" id="cancelFormBtn">Abbrechen</button>
                    <button type="submit" class="btn btn-primary">${isEdit ? 'Speichern' : 'Hinzufügen'}</button>
                </div>
            </form>`;
    },

    async submit(event, id = null) {
        event.preventDefault();

        const name = document.getElementById('lmName').value.trim();
        const anzahl = document.getElementById('lmAnzahl').value;
        const einheit = document.getElementById('lmEinheit').value;
        const lebensmittel_kategorie = document.getElementById('lmKategorie').value;
        const mahlzeiten = [...document.querySelectorAll('.lm-mahlzeit:checked')]
            .map(cb => cb.value);
        const gefrierfach = document.getElementById('lmGefrierfach')?.checked || false;

        let valid = true;
        document.getElementById('lmName').closest('.form-group').classList.toggle('error', !name);
        if (!name) valid = false;

        const anzahlValid = anzahl !== '' && Number(anzahl) >= 0;
        document.getElementById('lmAnzahl').closest('.form-group').classList.toggle('error', !anzahlValid);
        if (!anzahlValid) valid = false;

        if (!valid) {
            Toast.warning('Bitte füllen Sie alle Pflichtfelder aus');
            return;
        }

        const item = { name, anzahl, einheit, lebensmittel_kategorie, mahlzeiten, gefrierfach };
        try {
            if (id) {
                await DataManager.updateLebensmittel(id, item);
            } else {
                await DataManager.createLebensmittel(item);
            }
            Modal.close();
        } catch (error) {
            Toast.error(error.message);
        }
    },

    bind(id = null) {
        const form = document.getElementById('lebensmittelForm');
        form?.addEventListener('submit', (e) => this.submit(e, id));
        document.getElementById('cancelFormBtn')?.addEventListener('click', () => Modal.close());
    },

    open(id = null) {
        let item = null;
        if (id) {
            item = Store.getState().lebensmittel.find(l => l.id === id);
            if (!item) return;
        }
        Modal.open({
            title: id ? 'Lebensmittel bearbeiten' : 'Neues Lebensmittel',
            content: this.render(item),
            onClose: null
        });
        this.bind(id);
        setTimeout(() => document.getElementById('lmName')?.focus(), 50);
    }
};

// =========================================================================
// ALLGEMEINE BESTÄTIGUNGS-DIALOG
// =========================================================================
// UI.confirm zeigt einen Modal-Dialog mit Ja/Nein und löst ein Promise auf.
// @param {string} message - Bestätigungsnachricht (HTML wird escaped)
// @returns {Promise<boolean>} - true = bestätigt, false = abgebrochen
UI.confirm = async function(message) {
    return new Promise((resolve) => {
        Modal.open({
            title: 'Bestätigung erforderlich',
            content: `
                <div class="confirm-dialog">
                    <i class="fas fa-question-circle confirm-icon"></i>
                    <p>${UI.escapeHtml(message)}</p>
                </div>`,
            footer: `
                <button class="btn btn-outline" id="confirmCancelBtn">Abbrechen</button>
                <button class="btn btn-primary" id="confirmOkBtn"><i class="fas fa-check"></i> Ja, bestätigen</button>`
        });

        document.getElementById('confirmCancelBtn').addEventListener('click', () => {
            Modal.close();
            resolve(false);
        });

        document.getElementById('confirmOkBtn').addEventListener('click', () => {
            Modal.close();
            resolve(true);
        });
    });
};

// =========================================================================
// LÖSCH-BESTÄTIGUNG
// =========================================================================
async function confirmDelete(id) {
    const item = Store.getState().lebensmittel.find(l => l.id === id);
    const name = item ? item.name : 'dieses Lebensmittel';

    Modal.open({
        title: 'Löschen bestätigen',
        content: `
            <div class="confirm-delete">
                <i class="fas fa-exclamation-triangle confirm-icon"></i>
                <p>Möchten Sie <strong>${UI.escapeHtml(name)}</strong> wirklich löschen?</p>
                <p class="text-muted">Diese Aktion kann nicht rückgängig gemacht werden.</p>
            </div>`,
        footer: `
            <button class="btn btn-outline" id="cancelDeleteBtn">Abbrechen</button>
            <button class="btn btn-error" id="confirmDeleteBtn"><i class="fas fa-trash"></i> Löschen</button>`
    });

    document.getElementById('cancelDeleteBtn').addEventListener('click', () => Modal.close());
    document.getElementById('confirmDeleteBtn').addEventListener('click', async () => {
        Modal.close();
        try {
            await DataManager.deleteLebensmittel(id);
        } catch (error) {
            Toast.error(error.message);
        }
    });
}

// =========================================================================
// STARTUP - auf jeder Seite identisch
// =========================================================================
document.addEventListener('DOMContentLoaded', () => {
    Toast.init();
    Modal.init();
    Loading.init();
    Theme.init();
    Layout.init();

    // Seiten-spezifische Initialisierung, falls definiert
    if (typeof PageInit === 'function') {
        PageInit();
    }

    console.log(`KitchenOps geladen: ${location.pathname.split('/').pop() || 'index.html'}`);
});
