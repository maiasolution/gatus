# HANDOFF — Gatus Maia Solution Status Page

> Documento di handoff per continuare il lavoro in una nuova sessione.
> Ultimo aggiornamento: 2026-06-11

---

## 1. Stato attuale del progetto

### Completato ✅
- Fork di `TwiN/gatus` → `maiasolution/gatus` su GitHub
- Frontend Vue 3 rebrandizzato con Maia Solution:
  - `MaiaLogo.vue` — inline SVG con dark mode reattivo (MutationObserver)
  - Brand colors `#0066CC` / `#00AAFF` in `index.css`
  - Favicon light/dark (PNG Framer)
  - Fix titolo tab in dev mode (il parser Go template non vede `{{`)
- `EndpointCard.vue` — badge uptime SVG 7d/30d + indicatore SSL
- `EndpointDetails.vue` — statistiche uptime 24h/7d/30d, barra 30 giorni, grafico response time
- `Dockerfile` multi-stage: Node 20 (build Vue) → Go Alpine (binary + embed) → scratch
- `config.yaml` — dev locale con SQLite
- `config.prod.yaml` — produzione con PostgreSQL via `$GATUS_DB_URL`
- `docker-compose.dev.yml` — backend dev su porta 8080 con SQLite in volume Docker
- GitHub Actions CI/CD:
  - `build-push-dev.yml` — trigger su tag `dev*` → push a GAR
  - `build-push-prod.yml` — trigger su tag semver `X.Y.Z` → push a GAR
- Build locale testata e funzionante: `docker build -t maia-status:local .`
- **Feature maintenance events** (commit `81a4c1b8`) — finestre di manutenzione ad-hoc per endpoint via file JSON, hot-reloaded senza restart (vedi sezione 9)

### In corso / Pendente ⏳
- Deploy del container su Hetzner (vedi `TODO.md` per checklist dettagliata)
- Verifica endpoint Clack API (`/health` restituisce 403 — URL da correggere)
- Setup DNS `gatus.maiasolution.it` → IP Hetzner
- Reverse proxy (Caddy o Nginx) con TLS davanti alla porta 8080
- Sul server Hetzner: creare `/opt/gatus/maintenance-events.json` (anche vuoto `{}`) e montarlo nel container (vedi sezione 9)

---

## 2. Architettura

### Stack tecnologico
| Layer | Tecnologia |
|-------|-----------|
| Backend | Go (Fiber v2) — compilato in Docker, no Go locale |
| Frontend | Vue 3 + Vue CLI (non Vite) + Tailwind CSS |
| Storage dev | SQLite via volume Docker (`/data/gatus.db`) |
| Storage prod | PostgreSQL — connection string via `$GATUS_DB_URL` |
| Container registry | Google Artifact Registry (`europe-west8-docker.pkg.dev/maia-artifacts/public/maia-status`) |
| CI/CD | GitHub Actions + Workload Identity Federation (WIF) per GCP auth |

### Porte
| Servizio | Porta | Contesto |
|----------|-------|----------|
| Gatus backend (dev Docker) | 8080 | `docker compose -f docker-compose.dev.yml up -d` |
| Vue CLI dev server | 8081 | `npm run serve` in `web/app/` |
| Test container locale | 8090 | `docker run -p 8090:8080 maia-status:local` |

### Come funzionano insieme in dev
```
Browser :8081 → Vue CLI dev server
                    ↓ proxy /api/* /css/* /oicd
              Gatus Go backend :8080 (Docker)
```

### Frontend embed (produzione)
`npm run build` → `web/static/` → `go:embed static` in `web/static.go` → binario self-contained

### Go template rendering
`web/app/public/index.html` contiene variabili Go template (`{{ .UI.Header }}` ecc.) che vengono
rese da `api/spa.go` a runtime. Il Vue CLI dev server NON le renderizza → IIFE di fallback in
`index.html` rileva le variabili non renderizzate con `startsWith('{' + '{')` (il `{{` letterale
causerebbe un parse error del template Go).

---

## 3. File principali del progetto

### Configurazione
| File | Scopo |
|------|-------|
| `config.yaml` | Config dev — SQLite, endpoint Clack API e Clack Frontend |
| `config.prod.yaml` | Config produzione — PostgreSQL via `$GATUS_DB_URL`, stesso UI |
| `.dockerignore` | Esclude `web/app/node_modules` (non `web/app/` — serve per il build) |
| `.gitignore` | Esclude `.env`, `*.pem`, `*.key`, `*.db`, `node_modules` |

### Docker / CI
| File | Scopo |
|------|-------|
| `Dockerfile` | Multi-stage: frontend Node → backend Go → scratch runtime |
| `docker-compose.dev.yml` | Avvia Gatus dev su :8080 con image ufficiale + SQLite volume |
| `.github/workflows/build-push-dev.yml` | Build + push su tag `dev*` |
| `.github/workflows/build-push-prod.yml` | Build + push su tag semver (`1.0.0`) |

### Manutenzione eventi
| File | Scopo |
|------|-------|
| `maintenance-events.json` | File JSON con finestre di manutenzione ad-hoc per endpoint — modificabile a runtime senza restart |
| `config/maintenance/events.go` | **NUOVO** — EventsStore: legge il file, rileva modifiche via mtime, espone eventi al watchdog e all'API |
| `api/maintenance.go` | **NUOVO** — Handler `GET /api/v1/maintenance` |

### Frontend modificato (rispetto a TwiN/gatus upstream)
| File | Modifica |
|------|---------|
| `web/app/src/components/MaiaLogo.vue` | **NUOVO** — SVG inline Maia Solution, dark mode reattivo |
| `web/app/src/assets/logo-maia.svg` | **NUOVO** — SVG statico (per riferimento, il componente usa inline) |
| `web/app/src/App.vue` | Usa `<MaiaLogo>` invece di `<img>` logo, rimosso h1 header |
| `web/app/src/components/EndpointCard.vue` | Badge uptime 7d/30d + SSL indicator da `conditionResults` |
| `web/app/src/views/EndpointDetails.vue` | Uptime stats (24h/7d/30d), barra 30 giorni giornaliera, response time chart |
| `web/app/src/index.css` | Colori CSS var: `--primary: 210 100% 40%` (#0066CC), accent #00AAFF |
| `web/app/public/index.html` | Dev fallback IIFE, favicon light/dark, fix titolo tab |
| `web/app/src/App.vue` | Fetch `GET /api/v1/maintenance` al mount (refresh 5 min), `provide` dati ai componenti figli |
| `web/app/src/components/EndpointCard.vue` | Badge arancione (manutenzione attiva) / blu (pianificata entro 24h) |
| `web/app/src/views/EndpointDetails.vue` | Sezione "Scheduled Maintenance" con badge In corso / Pianificata / Completata |

### Documentazione
| File | Scopo |
|------|-------|
| `DEPLOY.md` | Guida completa: build immagine, setup PostgreSQL, deploy Hetzner, reverse proxy |
| `TODO.md` | Checklist con stato: ✅ completato / ⬜ pendente |
| `HANDOFF.md` | Questo file |

---

## 4. Variabili e segreti

### GitHub Actions — Secrets (Settings → Secrets → Actions)
| Nome | Descrizione |
|------|-------------|
| `WIF_PROVIDER` | Workload Identity Provider GCP (già configurato) |
| `WIF_SERVICE_ACCOUNT` | Service account GCP per push su Artifact Registry (già configurato) |

### GitHub Actions — env nel workflow (hardcoded in `.github/workflows/*.yml`)
| Nome | Valore |
|------|--------|
| `GAR_PROJECT_ID` | `maia-artifacts` |
| `GAR_LOCATION` | `europe-west8` |
| `REPOSITORY` | `public` |
| `SERVICE` | `maia-status` |

### Variabili d'ambiente runtime container (produzione)
| Nome | Descrizione |
|------|-------------|
| `GATUS_DB_URL` | Connection string PostgreSQL — da mettere in `/opt/gatus/.env` sul server (chmod 600) |
| `GATUS_CONFIG_PATH` | Default `/config/config.yaml` (già nel Dockerfile) |
| `GATUS_LOG_LEVEL` | Default `INFO` |

---

## 5. Git — Branch e remote

### Remote
```
origin    https://github.com/maiasolution/gatus.git   ← fork Maia Solution (push qui)
upstream  https://github.com/TwiN/gatus.git           ← progetto originale (solo fetch)
```

### Branch
| Branch | Scopo |
|--------|-------|
| `master` | Branch principale — tutti i cambiamenti vanno qui |

### Tagging per CI/CD
| Pattern tag | Workflow triggered | Immagine prodotta |
|-------------|-------------------|-------------------|
| `dev`, `dev-*` | `build-push-dev.yml` | `…/public/maia-status:dev` + `:SHA` |
| `1.0.0`, `1.2.3` | `build-push-prod.yml` | `…/public/maia-status:1.0.0` + `:SHA` + `:latest` |

### Aggiornare dall'upstream
```bash
git fetch upstream
git merge upstream/master
# risolvere eventuali conflitti su file modificati da Maia
```

---

## 6. Problemi aperti

| Problema | Dettaglio | Priorità |
|----------|-----------|----------|
| **Clack API 403** | `https://clackapi.maiasolution.it/health` risponde 403. L'URL dell'health check è probabilmente diverso. Da verificare e aggiornare in `config.yaml` e `config.prod.yaml` | Alta |
| **Deploy Hetzner** | Container mai deployato in prod. Vedi checklist in `TODO.md` | Alta |
| **DNS** | `gatus.maiasolution.it` non ancora puntato al server | Alta |
| **TLS / reverse proxy** | Caddy o Nginx da configurare davanti porta 8080 | Media |
| **PostgreSQL init** | Database e utente da creare sul server Hetzner (tabelle create in automatico da Gatus al primo avvio) | Media |

---

## 7. Comandi utili

### Sviluppo locale
```powershell
# Avviare backend dev (porta 8080)
docker compose -f docker-compose.dev.yml up -d

# Fermare backend dev
docker compose -f docker-compose.dev.yml down

# Avviare dev server Vue (porta 8081, hot-reload)
cd web\app
npm run serve
```

### Build e test container
```powershell
# Build immagine completa (lenta la prima volta, cache nelle successive)
docker build -t maia-status:local .

# Avviare container di test (porta 8090 per non conflitti con dev)
docker run -d --name maia-status-test -p 8090:8080 `
  -v "${PWD}/config.yaml:/config/config.yaml:ro" `
  -v maia-status-data:/data `
  maia-status:local

# Vedere log
docker logs maia-status-test

# Fermare e rimuovere
docker stop maia-status-test && docker rm maia-status-test
```

### Git e CI/CD
```bash
# Push con tag dev (trigger build-push-dev)
git tag dev-1 && git push origin dev-1

# Push con tag semver (trigger build-push-prod)
git tag 1.0.1 && git push origin 1.0.1

# Aggiornare da upstream TwiN/gatus
git fetch upstream && git merge upstream/master

# Eliminare e ricreare un tag
git tag -d dev && git push origin :dev
git tag dev && git push origin dev
```

### Immagine su GAR (dopo build CI riuscita)
```bash
# Formato immagine
# europe-west8-docker.pkg.dev/maia-artifacts/public/maia-status:TAG
```

---

## 8. Note importanti — Gotcha

### Go template e JavaScript
Il file `web/app/public/index.html` è un template Go renderizzato da `api/spa.go`.
**Non scrivere mai `{{` letteralmente dentro JavaScript** — il parser Go lo interpreta come
apertura di un'azione template. Per confrontare stringhe non-renderizzate usare:
```js
// ✅ Corretto
window.config.header.startsWith('{' + '{')

// ❌ Causa "unterminated character constant" al deploy
window.config.header.startsWith('{{')
```

### `.dockerignore` e `web/app/`
Il `.dockerignore` originale di TwiN escludeva `web/app` perché il vecchio Dockerfile usava
`web/static/` già committata. Il nostro Dockerfile multi-stage ha bisogno di `web/app/` per
buildare il frontend. Il `.dockerignore` ora esclude solo `web/app/node_modules`.

### Dev mode vs produzione
In dev (Vue CLI :8081) le variabili Go template (`{{ .UI.Header }}` ecc.) NON vengono renderizzate.
L'IIFE in `index.html` inietta i valori di fallback hardcoded. Ci sono **due posti** da aggiornare
se cambi logo/header: `config.yaml` (produzione) e l'IIFE in `index.html` (dev).

### SQLite — maximum-number-of-results
`maximum-number-of-results: 43200` = 60 check/ora × 24h × 30 giorni. Dimensionato per la barra
giornaliera a 30 giorni in `EndpointDetails.vue`. Se cambi `interval` in config, ricalcola:
`(3600 / seconds) × 24 × 30`.

### PostgreSQL — creazione tabelle
Gatus crea **automaticamente** tutte le tabelle al primo avvio (`CREATE TABLE IF NOT EXISTS`).
Devi creare manualmente solo il database e l'utente (vedi `DEPLOY.md`).

### Build GitHub Actions — versioni action
I workflow usano versioni fissate compatibili con il pattern funzionante:
`google-github-actions/auth@v1`, `docker/login-action@v1`, `docker/build-push-action@v2`.
**Non aggiornare** queste versioni senza testare — in passato versioni diverse usavano
output diversi (`access_token` vs `auth_token`).

### Workflow TwiN rimossi
I workflow originali `publish-*.yml` e `regenerate-static-assets.yml` sono stati **eliminati**
perché tentavano push su Docker Hub/GHCR con credenziali mancanti. Non ripristinarli.

### maintenance-events.json — formato e hot-reload
Il file è letto automaticamente dal watchdog ad ogni check cycle (ogni 60 secondi per default).
**Non è necessario riavviare il container** per applicare le modifiche.

**Struttura del file:**
```json
{
  "<endpoint-key>": [
    {
      "id": "identificativo-unico",
      "description": "Descrizione leggibile dell'evento",
      "start": "2026-06-15T22:00:00+02:00",
      "end":   "2026-06-16T00:00:00+02:00"
    }
  ]
}
```

**Campi:**
| Campo | Tipo | Obbligatorio | Note |
|-------|------|:---:|-------|
| `id` | string | ✅ | Identificativo libero, appare nei log |
| `description` | string | ✅ | Label mostrata nel frontend |
| `start` | RFC3339 datetime | ✅ | Includere sempre il timezone (`+02:00` per CEST) |
| `end` | RFC3339 datetime | ✅ | Includere sempre il timezone |

**Come si calcola l'endpoint key:**
`group` + `_` + `name`, tutto lowercase, spazi → `-`, `/` → `-`, `.` → `-`

Esempi:
| Group | Name | Key |
|-------|------|-----|
| Core | Clack API | `core_clack-api` |
| Core | Clack Frontend | `core_clack-frontend` |

**Effetti durante una finestra attiva:**
- I check vengono eseguiti normalmente (i dati di response time restano visibili)
- I contatori di uptime **non vengono aggiornati** → il downtime pianificato non abbassa la percentuale SLO
- Gli alert non vengono inviati (comportamento già esistente della `maintenance` globale di Gatus)
- Frontend: badge arancione "Manutenzione fino alle HH:MM" sulla card

**File vuoto (nessuna manutenzione attiva):**
```json
{}
```

**Deploy su Hetzner — aggiungere al comando `docker run`:**
```bash
-v /opt/gatus/maintenance-events.json:/config/maintenance-events.json:ro
```
Il file `config.prod.yaml` ha già la riga `maintenance-events-file: /config/maintenance-events.json`.
Creare il file sul server prima del primo avvio: `echo '{}' > /opt/gatus/maintenance-events.json`

---

### MaiaLogo dark mode
`MaiaLogo.vue` usa un `MutationObserver` su `document.documentElement` per rilevare il toggle
`.dark` di Gatus. Il fill del testo cambia: `#141D3A` (light) → `#F8FAFC` (dark).
I colori chevron/puntini sono fissi.
