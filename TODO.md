# Gatus — Maia Solution TODO

## Completato

### Setup e sviluppo
- [x] Fork TwiN/gatus → [maiasolution/gatus](https://github.com/maiasolution/gatus)
- [x] Remote `origin` = fork Maia Solution, `upstream` = TwiN/gatus
- [x] `docker-compose.dev.yml` — backend su :8080 con SQLite in volume Docker
- [x] Dev server Vue CLI su :8081 con proxy `/api` → :8080

### Frontend / brand
- [x] Colori brand in `index.css` — primario `#0066CC`, accent `#00AAFF`
- [x] `MaiaLogo.vue` — inline SVG con dark mode reattivo (MutationObserver su `.dark`)
- [x] `App.vue` — logo sostituito con `<MaiaLogo>`, header title rimosso (nel SVG)
- [x] `index.html` — favicon light/dark (Framer), fix titolo tab in dev mode
- [x] `EndpointCard.vue` — badge uptime SVG 7d/30d + indicatore SSL da conditionResults
- [x] `EndpointDetails.vue` — statistiche uptime (24h/7d/30d), barra giornaliera 30gg, grafico response time

### Infrastruttura
- [x] `Dockerfile` — multi-stage: Node (build Vue) → Go (binary + embed) → scratch
- [x] `config.yaml` — dev con SQLite, endpoint Clack API e Clack Frontend
- [x] `config.prod.yaml` — PostgreSQL via `$GATUS_DB_URL` (nessuna credenziale nel file)
- [x] `.gitignore` — aggiunto `.env`, `*.pem`, `*.key`
- [x] `DEPLOY.md` — guida completa: registry, PostgreSQL, Hetzner, reverse proxy

---

## Da fare

### Deploy su Hetzner

- [ ] **Build e push immagine** al registry privato
  ```powershell
  docker build -t YOUR_REGISTRY/maia-gatus:latest .
  docker push YOUR_REGISTRY/maia-gatus:latest
  ```

- [ ] **Creare database PostgreSQL** su Hetzner (vedi `DEPLOY.md` → sezione PostgreSQL)
  ```sql
  CREATE DATABASE gatus;
  CREATE USER gatus_user WITH PASSWORD '...';
  GRANT ALL PRIVILEGES ON DATABASE gatus TO gatus_user;
  ```

- [ ] **Copiare `config.prod.yaml`** sul server come `/opt/gatus/config.yaml`

- [ ] **Creare `/opt/gatus/.env`** con `GATUS_DB_URL=postgres://...` (`chmod 600`)

- [ ] **Avviare il container** sul server Hetzner (vedi `DEPLOY.md` → sezione Deploy)

- [ ] **DNS** — puntare `gatus.maiasolution.it` all'IP del server Hetzner

- [ ] **Reverse proxy + TLS** — configurare Caddy o Nginx davanti a :8080

### Verifica post-deploy

- [ ] **Health endpoint Clack API** — `clackapi.maiasolution.it/health` restituisce attualmente 403.
  Verificare l'endpoint corretto (potrebbe essere `/`, `/status`, o richiedere header specifici)
  e aggiornare `config.yaml` / `config.prod.yaml`

- [ ] **Test iframe Framer** — verificare che `https://gatus.maiasolution.it` sia embeddabile
  senza errori `X-Frame-Options` / CSP dal reverse proxy

---

## Opzionale / Futuro

- [ ] **GitHub Actions CI/CD** — build + push automatico al registry su ogni push a `master`
- [ ] **Alert** — configurare notifiche (email/Slack/webhook) quando un endpoint va down
- [ ] **Nuovi endpoint** — aggiungere altri servizi Maia Solution da monitorare
- [ ] **Merge upstream** — quando TwiN rilascia aggiornamenti: `git fetch upstream && git merge upstream/master`
