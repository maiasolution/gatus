# Deploy su Hetzner

## Prerequisiti
- Docker installato localmente e sul server Hetzner
- Accesso SSH al server Hetzner
- Istanza PostgreSQL già attiva su Hetzner (o accessibile dalla rete del server)

---

## Sviluppo locale

**Backend (Go) — nessun Go richiesto localmente:**
```powershell
docker compose -f docker-compose.dev.yml up -d
```

**Frontend Vue (hot-reload su porta 8081):**
```powershell
cd web\app
npm run serve
```

Apri http://localhost:8081. Le chiamate `/api/*` vengono proxate automaticamente verso http://localhost:8080.

> Config usata in dev: `config.yaml` (SQLite locale in Docker volume)

---

## Build e push immagine

Il Dockerfile è multi-stage: compila il frontend Vue con Node, poi il binario Go con il frontend embeddato, senza richiedere niente localmente tranne Docker.

```powershell
# Build
docker build -t YOUR_REGISTRY/maia-gatus:latest .

# Tag con versione (consigliato)
docker build -t YOUR_REGISTRY/maia-gatus:1.0.0 .
docker tag YOUR_REGISTRY/maia-gatus:1.0.0 YOUR_REGISTRY/maia-gatus:latest

# Login al registry (se necessario)
docker login YOUR_REGISTRY

# Push
docker push YOUR_REGISTRY/maia-gatus:latest
docker push YOUR_REGISTRY/maia-gatus:1.0.0
```

> Sostituisci `YOUR_REGISTRY` con l'indirizzo del tuo registry privato (es. `registry.hetzner.com/maiasolution` o `ghcr.io/maiasolution`).

---

## Configurazione PostgreSQL su Hetzner

### Prepara il database

Sul server Hetzner (o sul host Postgres), crea database e utente dedicati:

```sql
CREATE DATABASE gatus;
CREATE USER gatus_user WITH PASSWORD 'scegli-una-password-sicura';
GRANT ALL PRIVILEGES ON DATABASE gatus TO gatus_user;
-- Su PostgreSQL 15+ serve anche:
GRANT ALL ON SCHEMA public TO gatus_user;
```

### Connection string

Il formato è:
```
postgres://gatus_user:PASSWORD@HOST:5432/gatus?sslmode=require
```

- Se Postgres è sullo stesso server Hetzner: usa `localhost` o `127.0.0.1`
- Se è su un DB managed Hetzner separato: usa l'host fornito dal pannello Hetzner
- `sslmode=require` è raccomandato su Hetzner managed DB; usa `sslmode=disable` solo se la connessione è interna alla stessa macchina

---

## Deploy su Hetzner

### 1. Prepara le directory e il file di config

```bash
ssh user@HETZNER_IP

mkdir -p /opt/gatus
```

Copia il file di config produzione:
```bash
# Da locale:
scp config.prod.yaml user@HETZNER_IP:/opt/gatus/config.yaml
```

### 2. Crea il file delle variabili d'ambiente

```bash
# Sul server Hetzner
cat > /opt/gatus/.env <<EOF
GATUS_DB_URL=postgres://gatus_user:PASSWORD@HOST:5432/gatus?sslmode=require
EOF

chmod 600 /opt/gatus/.env
```

### 3. Pull e avvio container

```bash
docker login YOUR_REGISTRY   # se il registry è privato

docker pull YOUR_REGISTRY/maia-gatus:latest

docker run -d \
  --name gatus \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /opt/gatus/config.yaml:/config/config.yaml:ro \
  --env-file /opt/gatus/.env \
  YOUR_REGISTRY/maia-gatus:latest
```

> Nessun volume `/data` necessario: PostgreSQL gestisce la persistenza.

---

## Aggiornamenti

### Solo config (senza rebuild immagine)

```bash
scp config.prod.yaml user@HETZNER_IP:/opt/gatus/config.yaml
ssh user@HETZNER_IP "docker restart gatus"
```

### Nuova versione frontend o backend

```powershell
# Rebuild e push da locale
docker build -t YOUR_REGISTRY/maia-gatus:latest .
docker push YOUR_REGISTRY/maia-gatus:latest
```

```bash
# Sul server Hetzner
docker pull YOUR_REGISTRY/maia-gatus:latest
docker stop gatus && docker rm gatus
docker run -d \
  --name gatus \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /opt/gatus/config.yaml:/config/config.yaml:ro \
  --env-file /opt/gatus/.env \
  YOUR_REGISTRY/maia-gatus:latest
```

---

## Reverse proxy (Nginx/Caddy)

Gatus gira su `:8080`. Metti un reverse proxy davanti per HTTPS.

**Caddy** (consigliato — gestisce TLS automaticamente):
```
gatus.maiasolution.it {
    reverse_proxy localhost:8080
}
```

**Nginx:**
```nginx
server {
    listen 443 ssl;
    server_name gatus.maiasolution.it;
    # ssl_certificate / ssl_certificate_key ...

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

> Non aggiungere header `X-Frame-Options` se vuoi supportare l'embedding in iframe (es. Framer).

---

## Embedding in Framer (iframe)

Gatus non imposta `X-Frame-Options`, quindi funziona di default. In Framer usa un componente **Embed**:

```html
<iframe src="https://gatus.maiasolution.it" width="100%" height="600" frameborder="0"></iframe>
```

---

## Struttura directory sul server

```
/opt/gatus/
├── config.yaml    # config produzione (montata come volume read-only)
└── .env           # variabili d'ambiente con credenziali DB (chmod 600)
```
