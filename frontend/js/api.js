/**
 * KitchenOps API Client
 * Handhabt alle Kommunikationen mit dem Go-Backend.
 *
 * Basis-URL-Auflösung:
 * - Dev (Live Server / python -m http.server auf Port 5500x): http://localhost:3000
 * - Docker/Produktion (nginx auf Port 80/443, API über Reverse-Proxy): gleicher Host
 * - Manuell überschreibbar über die Einstellungen (localStorage) oder setBaseUrl
 */

// defaultBaseUrl bestimmt die Backend-URL ohne gespeicherte Einstellung.
function defaultBaseUrl() {
    // Entwicklungsserver auf Port 5500-5509: Backend separat auf localhost:3000
    if (/^550[0-9]$/.test(location.port)) {
        return 'http://localhost:3000';
    }
    // Produktion (nginx, Port 80/443): API läuft hinter dem Reverse-Proxy
    // auf demselben Host - relative Anfragen, kein CORS nötig.
    return '';
}

class KitchenOpsAPI {
    constructor() {
        this.baseUrl = localStorage.getItem('apiBaseUrl') || defaultBaseUrl();
        this.saveBaseUrl(this.baseUrl);
    }

    /**
     * Setzt die Base URL für API-Anfragen
     * @param {string} url - Neue Base URL
     */
    setBaseUrl(url) {
        this.baseUrl = url;
        this.saveBaseUrl(url);
    }

    /**
     * Speichert die Base URL im LocalStorage
     * @param {string} url - Base URL
     */
    saveBaseUrl(url) {
        localStorage.setItem('apiBaseUrl', url);
    }

    /**
     * Führt eine HTTP-Anfrage aus
     * @param {string} endpoint - API-Endpoint (z.B. '/lebensmittel')
     * @param {Object} options - Fetch-Optionen
     * @returns {Promise<Response>} - Response-Objekt
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        
        try {
            const response = await fetch(url, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });

            if (!response.ok) {
                const errorData = await this.parseResponse(response);
                throw new Error(
                    errorData?.message || 
                    errorData?.error || 
                    `HTTP Error: ${response.status} ${response.statusText}`
                );
            }

            return this.parseResponse(response);
        } catch (error) {
            if (error.name === 'TypeError' && error.message.includes('Failed to fetch')) {
                throw new Error('Verbindung zum Server fehlgeschlagen. Bitte überprüfen Sie die Server-URL.');
            }
            throw error;
        }
    }

    /**
     * Parsed die Response basierend auf dem Content-Type
     * @param {Response} response - Fetch Response
     * @returns {Promise<any>} - Geparsedes Ergebnis
     */
    async parseResponse(response) {
        const contentType = response.headers.get('content-type');
        
        if (contentType && contentType.includes('application/json')) {
            return response.json();
        } else if (response.status === 204) {
            return null; // No Content
        } else {
            return response.text();
        }
    }

    // =========================================================================
    // LEBENSMITTEL ENDPOINTS
    // =========================================================================

    /**
     * Holt alle Lebensmittel
     * @returns {Promise<Array>} - Liste der Lebensmittel
     */
    async getAllLebensmittel() {
        return this.request('/lebensmittel');
    }

    /**
     * Holt ein einzelnes Lebensmittel nach ID
     * @param {number} id - Lebensmittel-ID
     * @returns {Promise<Object>} - Lebensmittel-Objekt
     */
    async getLebensmittelById(id) {
        return this.request(`/lebensmittel/${id}`);
    }

    /**
     * Erstellt ein neues Lebensmittel
     * @param {Object} lebensmittel - Lebensmittel-Daten
     * @param {string} lebensmittel.name - Name des Lebensmittels
     * @param {number} lebensmittel.anzahl - Menge
     * @param {string} lebensmittel.einheit - Einheit (z.B. "g", "kg", "Stück")
     * @returns {Promise<Object>} - Erstelltes Lebensmittel
     */
    async createLebensmittel(lebensmittel) {
        return this.request('/lebensmittel', {
            method: 'POST',
            body: JSON.stringify(lebensmittel)
        });
    }

    /**
     * Aktualisiert ein Lebensmittel
     * @param {number} id - Lebensmittel-ID
     * @param {Object} lebensmittel - Aktualisierte Daten
     * @returns {Promise<Object>} - Aktualisiertes Lebensmittel
     */
    async updateLebensmittel(id, lebensmittel) {
        return this.request(`/lebensmittel/${id}`, {
            method: 'PUT',
            body: JSON.stringify(lebensmittel)
        });
    }

    /**
     * Löscht ein Lebensmittel
     * @param {number} id - Lebensmittel-ID
     * @returns {Promise<Object>} - Lösch-Bestätigung
     */
    async deleteLebensmittel(id) {
        return this.request(`/lebensmittel/${id}`, {
            method: 'DELETE'
        });
    }

    /**
     * Kategorisiert alle noch nicht zugeordneten Lebensmittel:
     * per Regeln und verbleibende per KI (ein Batch-Aufruf)
     * @returns {Promise<Object>} - {success, message, per_regeln, per_llm, offen}
     */
    async kategorisierenLebensmittel() {
        return this.request('/lebensmittel/kategorisieren', {
            method: 'POST'
        });
    }

    /**
     * Erfasst Lebensmittel per Freitext über die KI
     * (z.B. "Füge Brot, Butter und zwei mal Sahne hinzu")
     * @param {string} text - Freitext
     * @returns {Promise<Object>} - {lebensmittel: [...]} der erfassten Artikel
     */
    async erfasseLebensmittel(text) {
        return this.request('/lebensmittel/erfassen', {
            method: 'POST',
            body: JSON.stringify({ text })
        });
    }

    /**
     * Erfasst Einkaufsartikel per Freitext über die KI
     * (z.B. "Füge mir Milch, Eier und zwei Packungen Mehl hinzu")
     * @param {string} text - Freitext
     * @returns {Promise<Object>} - {artikel: [...]} der erfassten Artikel
     */
    async erfasseEinkaufsartikel(text) {
        return this.request('/einkaufsliste/erfassen', {
            method: 'POST',
            body: JSON.stringify({ text })
        });
    }

    // =========================================================================
    // EINKAUFSLISTE ENDPOINTS
    // =========================================================================

    /**
     * Holt alle Einkaufslisten-Artikel
     * @returns {Promise<Object>} - {artikel: [{id, name, anzahl, einheit, notiz, erledigt}]}
     */
    async getAllEinkaufsartikel() {
        return this.request('/einkaufsliste');
    }

    /**
     * Fügt einen Artikel zur Einkaufsliste hinzu.
     * Existiert bereits ein offener Eintrag mit gleichem Namen, aktualisiert das Backend diesen.
     * @param {Object} artikel - {name, anzahl, einheit, notiz}
     * @returns {Promise<Object>} - Angelegter/aktualisierter Artikel
     */
    async createEinkaufsartikel(artikel) {
        return this.request('/einkaufsliste', {
            method: 'POST',
            body: JSON.stringify(artikel)
        });
    }

    /**
     * Aktualisiert einen Einkaufslisten-Artikel (auch zum Abhaken via erledigt: true)
     * @param {number} id - Artikel-ID
     * @param {Object} artikel - {name, anzahl, einheit, notiz, erledigt}
     * @returns {Promise<Object>} - Aktualisierter Artikel
     */
    async updateEinkaufsartikel(id, artikel) {
        return this.request(`/einkaufsliste/${id}`, {
            method: 'PUT',
            body: JSON.stringify(artikel)
        });
    }

    /**
     * Löscht einen einzelnen Einkaufslisten-Artikel
     * @param {number} id - Artikel-ID
     * @returns {Promise<Object>} - {success, message}
     */
    async deleteEinkaufsartikel(id) {
        return this.request(`/einkaufsliste/${id}`, {
            method: 'DELETE'
        });
    }

    /**
     * Leert die komplette Einkaufsliste
     * @returns {Promise<Object>} - {success, message}
     */
    async clearEinkaufsliste() {
        return this.request('/einkaufsliste', {
            method: 'DELETE'
        });
    }

    // =========================================================================
    // REZEPT ENDPOINTS
    // =========================================================================

    /**
     * Generiert Rezepte per KI auf Basis des Inventars
     * @param {Object} anfrage - {prompt, essens_kategorie, lebensmittel_kategorie, anzahl}
     * @returns {Promise<Object>} - {rezepte: [{name, beschreibung, essens_kategorie, ...}]}
     */
    async generateRezepte(anfrage) {
        return this.request('/rezepte/generieren', {
            method: 'POST',
            body: JSON.stringify(anfrage)
        });
    }

    /**
     * Speichert ein Rezept mit Verknüpfung auf Inventar-IDs
     * @param {Object} rezept - {name, lebensmittel_kategorie, essens_kategorie, zutaten_ids}
     * @returns {Promise<Object>} - {success, message, id}
     */
    async saveRezept(rezept) {
        return this.request('/rezepte', {
            method: 'POST',
            body: JSON.stringify(rezept)
        });
    }

    /**
     * Holt alle gespeicherten Rezepte inkl. Zutaten
     * @returns {Promise<Object>} - {rezepte: [{id, name, lebensmittel_kategorie, essens_kategorie, zutaten, created_at}]}
     */
    async getAllRezepte() {
        return this.request('/rezepte');
    }

    /**
     * Löscht ein gespeichertes Rezept
     * @param {number} id - Rezept-ID
     * @returns {Promise<Object>} - {success, message}
     */
    async deleteRezept(id) {
        return this.request(`/rezepte/${id}`, {
            method: 'DELETE'
        });
    }

    // =========================================================================
    // HILFSFUNKTIONEN
    // =========================================================================

    /**
     * Ping den Server um die Verbindung zu testen.
     * Nutzt bewusst eine echte API-Route (GET /lebensmittel) statt der
     * Banner-Seite, damit auch der Reverse-Proxy-Pfad geprüft wird.
     * @returns {Promise<boolean>} - true wenn Server erreichbar
     */
    async ping() {
        try {
            await this.request('/lebensmittel');
            return true;
        } catch (error) {
            return false;
        }
    }

    /**
     * Ändert die Server-URL und testet die Verbindung
     * @param {string} url - Neue Server-URL
     * @returns {Promise<boolean>} - true wenn Verbindung erfolgreich
     */
    async changeServerUrl(url) {
        this.setBaseUrl(url);
        return this.ping();
    }

    /**
     * Gibt die aktuelle Base URL zurück
     * @returns {string} - Aktuelle Base URL
     */
    getCurrentBaseUrl() {
        return this.baseUrl;
    }
}

// Singelton-Instanz
const api = new KitchenOpsAPI();
