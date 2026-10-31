/**
 * KitchenOps - Einstellungen-Seite
 * Server-URL, Theme, Bestands-Schwellenwerte, lokale Daten.
 */

function PageInit() {
    initServerSection();
    initThemeSection();
    initThresholdsSection();
    initDataSection();

    // Verbindung beim Laden prüfen (ohne Toast-Meldung)
    checkConnection(false);
}

// -------------------------------------------------------------------------
// Server-Verbindung
// -------------------------------------------------------------------------
function initServerSection() {
    const input = document.getElementById('serverUrlInput');
    input.value = api.getCurrentBaseUrl();

    document.getElementById('saveServerBtn').addEventListener('click', async () => {
        const url = input.value.trim().replace(/\/+$/, '');
        if (!url) {
            Toast.warning('Bitte geben Sie eine URL ein');
            return;
        }
        api.setBaseUrl(url);
        Toast.success('Server-URL gespeichert');
        await checkConnection(true);
    });

    document.getElementById('testConnectionBtn').addEventListener('click', async () => {
        const url = input.value.trim().replace(/\/+$/, '');
        if (url) api.setBaseUrl(url);
        await checkConnection(true);
    });
}

async function checkConnection(showToast) {
    const badge = document.getElementById('connectionStatus');

    badge.className = 'badge badge-info';
    badge.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Prüfen...';

    // ping() gibt true/false zurück und wirft KEINE Fehler -
    // der Rückgabewert muss ausgewertet werden, sonst zeigt der Test
    // auch ohne Verbindung fälschlich "Verbunden" an.
    let erreichbar = false;
    try {
        erreichbar = await api.ping();
    } catch (error) {
        erreichbar = false;
    }

    if (erreichbar) {
        Store.setState({ serverConnected: true });
        badge.className = 'badge badge-success';
        badge.innerHTML = '<i class="fas fa-circle-check"></i> Verbunden';
        if (showToast) Toast.success('Verbindung erfolgreich');
    } else {
        Store.setState({ serverConnected: false });
        badge.className = 'badge badge-error';
        badge.innerHTML = '<i class="fas fa-circle-xmark"></i> Getrennt';
        if (showToast) Toast.error('Verbindung fehlgeschlagen');
    }
}

// -------------------------------------------------------------------------
// Theme
// -------------------------------------------------------------------------
function initThemeSection() {
    const saved = localStorage.getItem('theme') || 'light';

    const options = document.querySelectorAll('.theme-option');
    options.forEach(btn => {
        btn.addEventListener('click', () => {
            const theme = btn.dataset.theme;

            if (theme === 'system') {
                const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
                localStorage.setItem('theme', prefersDark ? 'dark' : 'light');
                Theme.apply(prefersDark ? 'dark' : 'light');
                localStorage.setItem('themeSystem', '1');
                Toast.info('Design folgt dem System');
            } else {
                localStorage.setItem('theme', theme);
                localStorage.removeItem('themeSystem');
                Theme.apply(theme);
                Toast.success(theme === 'dark' ? 'Dunkles Design aktiviert' : 'Helles Design aktiviert');
            }
            updateThemeSelection();
        });
    });

    updateThemeSelection();

    // Auf System-Änderungen reagieren, wenn "System" gewählt ist
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (localStorage.getItem('themeSystem')) {
            Theme.apply(e.matches ? 'dark' : 'light');
        }
    });
}

function updateThemeSelection() {
    const saved = localStorage.getItem('theme') || 'light';
    const isSystem = !!localStorage.getItem('themeSystem');

    document.querySelectorAll('.theme-option').forEach(btn => {
        const selected =
            (isSystem && btn.dataset.theme === 'system') ||
            (!isSystem && btn.dataset.theme === saved);
        btn.classList.toggle('selected', selected);
    });
}

// -------------------------------------------------------------------------
// Bestands-Schwellenwerte
// -------------------------------------------------------------------------
function initThresholdsSection() {
    const lowInput = document.getElementById('lowThresholdInput');
    const criticalInput = document.getElementById('criticalThresholdInput');

    lowInput.value = localStorage.getItem('lowStockThreshold') || CONSTANTS.LOW_STOCK_THRESHOLD;
    criticalInput.value = localStorage.getItem('criticalStockThreshold') || CONSTANTS.CRITICAL_STOCK_THRESHOLD;

    document.getElementById('saveThresholdsBtn').addEventListener('click', () => {
        const low = parseInt(lowInput.value, 10);
        const critical = parseInt(criticalInput.value, 10);

        if (isNaN(low) || low < 1) {
            Toast.warning('Ungültiger Wert für "Niedriger Bestand"');
            return;
        }
        if (isNaN(critical) || critical < 0) {
            Toast.warning('Ungültiger Wert für "Kritischer Bestand"');
            return;
        }
        if (critical >= low) {
            Toast.warning('"Kritisch" muss kleiner als "Niedrig" sein');
            return;
        }

        localStorage.setItem('lowStockThreshold', String(low));
        localStorage.setItem('criticalStockThreshold', String(critical));

        // Laufende Konstanten aktualisieren
        CONSTANTS.LOW_STOCK_THRESHOLD = low;
        CONSTANTS.CRITICAL_STOCK_THRESHOLD = critical;

        Toast.success('Schwellenwerte gespeichert');
    });
}

// -------------------------------------------------------------------------
// Daten
// -------------------------------------------------------------------------
function initDataSection() {
    document.getElementById('reloadDataBtn').addEventListener('click', async () => {
        await DataManager.loadLebensmittel();
        Toast.success('Daten neu geladen');
    });

    document.getElementById('clearLocalDataBtn').addEventListener('click', () => {
        Modal.open({
            title: 'Lokale Daten löschen',
            content: `
                <div class="confirm-delete">
                    <i class="fas fa-exclamation-triangle confirm-icon"></i>
                    <p>Alle lokal gespeicherten Daten werden gelöscht:</p>
                    <ul class="data-list">
                        <li>Einkaufsliste</li>
                        <li>Theme-Einstellungen</li>
                        <li>Schwellenwerte</li>
                        <li>Ansichtseinstellungen</li>
                    </ul>
                    <p class="text-muted">Server-Daten bleiben unberührt.</p>
                </div>`,
            footer: `
                <button class="btn btn-outline" id="cancelClearBtn">Abbrechen</button>
                <button class="btn btn-error" id="confirmClearBtn"><i class="fas fa-broom"></i> Löschen</button>
            `
        });

        document.getElementById('cancelClearBtn').addEventListener('click', () => Modal.close());
        document.getElementById('confirmClearBtn').addEventListener('click', () => {
            [
                'kitchenops_einkaufsliste',
                'theme',
                'themeSystem',
                'lowStockThreshold',
                'criticalStockThreshold',
                'lmView'
            ].forEach(key => localStorage.removeItem(key));

            Modal.close();
            Toast.success('Lokale Daten gelöscht - Seite wird neu geladen');
            setTimeout(() => location.reload(), 800);
        });
    });
}
