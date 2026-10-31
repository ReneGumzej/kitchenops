/**
 * KitchenOps - Inventar-Seite
 * Gruppiert alle Lebensmittel nach Kategorien und zeigt Statistiken je Gruppe.
 */

let activeKategorie = 'all';

// Kategorie-Mapping (Einheit → Kategorie) + Farben und Icons
const KATEGORIEN = [
    { name: 'Trockenware', einheiten: ['g', 'kg', 'Packung', 'Beutel'], icon: 'fa-wheat', color: '#FFB74D' },
    { name: 'Flüssiges', einheiten: ['l', 'ml', 'Flasche'], icon: 'fa-droplet', color: '#2196F3' },
    { name: 'Konserven', einheiten: ['Dose', 'Glas'], icon: 'fa-jar', color: '#8E44AD' },
    { name: 'Stückware', einheiten: ['Stück', 'mg'], icon: 'fa-cube', color: '#4ECDC4' }
];

function PageInit() {
    bindEvents();
    DataManager.loadLebensmittel().then(render);
    Store.subscribe(() => render());
}

function getKategorie(einheit) {
    const k = KATEGORIEN.find(k => k.einheiten.includes(einheit));
    return k ? k.name : 'Sonstiges';
}

function getKategorieIcon(name) {
    if (name === 'Gefrierfach') {
        return { icon: 'fa-snowflake', color: '#2196F3' };
    }
    const k = KATEGORIEN.find(k => k.name === name);
    return k ? { icon: k.icon, color: k.color } : { icon: 'fa-box', color: '#9E9E9E' };
}

function bindEvents() {
    document.getElementById('refreshInventarBtn').addEventListener('click', async () => {
        await DataManager.loadLebensmittel();
        render();
    });

    // Kategorie-Chips
    document.getElementById('kategorieChips').addEventListener('click', (e) => {
        const chip = e.target.closest('.chip');
        if (!chip) return;

        activeKategorie = chip.dataset.kategorie;
        document.querySelectorAll('#kategorieChips .chip').forEach(c => c.classList.remove('active'));
        chip.classList.add('active');
        render();
    });

    // Bearbeiten/Löschen innerhalb der Gruppen
    document.getElementById('inventarContent').addEventListener('click', (e) => {
        const editBtn = e.target.closest('.edit-btn');
        const deleteBtn = e.target.closest('.delete-btn');
        const card = e.target.closest('.lebensmittel-card');

        if (editBtn) {
            e.stopPropagation();
            LebensmittelForm.open(parseInt(editBtn.dataset.id, 10));
        } else if (deleteBtn) {
            e.stopPropagation();
            confirmDelete(parseInt(deleteBtn.dataset.id, 10));
        } else if (card && !e.target.closest('.lebensmittel-actions')) {
            location.href = `lebensmittel.html?id=${card.dataset.id}`;
        }
    });
}

// -------------------------------------------------------------------------
// Rendering
// -------------------------------------------------------------------------
function render() {
    const { lebensmittel } = Store.getState();
    renderKategorieChips(lebensmittel);
    renderInventar(lebensmittel);
}

function renderKategorieChips(lebensmittel) {
    // Gruppieren - Gefrierfach-Artikel bilden eine eigene Gruppe
    const groups = {};
    lebensmittel.forEach(l => {
        const k = l.gefrierfach ? 'Gefrierfach' : getKategorie(l.einheit);
        groups[k] = (groups[k] || 0) + 1;
    });

    const total = lebensmittel.length;

    const gefrierfachCount = groups['Gefrierfach'] || 0;
    const gefrierfachChip = `
        <button class="chip ${activeKategorie === 'Gefrierfach' ? 'active' : ''}" data-kategorie="Gefrierfach">
            <i class="fas fa-snowflake"></i> Gefrierfach (${gefrierfachCount})
        </button>`;

    document.getElementById('kategorieChips').innerHTML = `
        <button class="chip ${activeKategorie === 'all' ? 'active' : ''}" data-kategorie="all">
            <i class="fas fa-border-all"></i> Alle (${total})
        </button>
        ${gefrierfachChip}
        ${Object.entries(groups)
            .filter(([name]) => name !== 'Gefrierfach')
            .sort((a, b) => b[1] - a[1])
            .map(([name, count]) => `
                <button class="chip ${activeKategorie === name ? 'active' : ''}" data-kategorie="${name}">
                    ${name} (${count})
                </button>`)
            .join('')}
    `;
}

function renderInventar(lebensmittel) {
    const container = document.getElementById('inventarContent');

    if (lebensmittel.length === 0) {
        container.innerHTML = UI.emptyState(
            'fa-clipboard-list',
            'Inventar ist leer',
            'Es wurden noch keine Lebensmittel angelegt.',
            `<a href="lebensmittel.html" class="btn btn-primary mt-md"><i class="fas fa-plus"></i> Jetzt anlegen</a>`
        );
        return;
    }

    // Nach Kategorie gruppieren - Gefrierfach-Artikel in einer eigenen Gruppe
    const grouped = {};
    lebensmittel.forEach(l => {
        const k = l.gefrierfach ? 'Gefrierfach' : getKategorie(l.einheit);
        if (!grouped[k]) grouped[k] = [];
        grouped[k].push(l);
    });

    // Gefrierfach-Gruppe zuerst anzeigen (Kühltruhe-Bestand prominent)
    const namen = Object.keys(grouped);
    namen.sort((a, b) => {
        if (a === 'Gefrierfach') return -1;
        if (b === 'Gefrierfach') return 1;
        return a.localeCompare(b, 'de');
    });

    // Filter anwenden
    const entries = activeKategorie === 'all'
        ? namen.map(name => [name, grouped[name]])
        : namen.filter(name => name === activeKategorie).map(name => [name, grouped[name]]);

    if (entries.length === 0) {
        container.innerHTML = UI.emptyState(
            'fa-folder-open',
            'Keine Artikel in dieser Kategorie',
            'Wählen Sie eine andere Kategorie.'
        );
        return;
    }

    container.innerHTML = entries.map(([kategorie, items]) => {
        const { icon, color } = getKategorieIcon(kategorie);
        const totalAmount = items.reduce((sum, l) => sum + l.anzahl, 0);
        const lowCount = items.filter(l => l.anzahl <= CONSTANTS.LOW_STOCK_THRESHOLD).length;

        return `
            <div class="card mb-lg">
                <div class="card-header">
                    <h3 class="kategorie-title">
                        <span class="kategorie-icon" style="background: ${color}22; color: ${color}">
                            <i class="fas ${icon}"></i>
                        </span>
                        ${kategorie}
                    </h3>
                    <div class="kategorie-stats">
                        <span class="badge badge-primary">${items.length} Artikel</span>
                        <span class="badge ${lowCount > 0 ? 'badge-warning' : 'badge-success'}">
                            ${lowCount > 0 ? `${lowCount} niedrig` : 'Bestand OK'}
                        </span>
                    </div>
                </div>
                <div class="card-body">
                    <div class="lebensmittel-grid">
                        ${items
                            .sort((a, b) => a.name.localeCompare(b.name, 'de'))
                            .map(l => UI.lebensmittelCard(l, true))
                            .join('')}
                    </div>
                </div>
            </div>`;
    }).join('');
}
