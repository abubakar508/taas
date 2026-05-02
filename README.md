# TAAS – Telecom-as-a-Service

TAAS is a multi-tenant telecom-as-a-service API built in Go using Fiber and PostgreSQL. It currently supports **Vodacom DRC airtime purchase** and **Vodacom DRC bundle offers and bundle purchase**, with more providers and services planned as the project grows.

This project is also open to **open-source contributions**. What exists today is production-oriented scaffolding for telecom integrations, and more countries, providers, and transaction types are expected to be added over time.

---

## Current scope

At the moment, TaaS supports the following Vodacom DRC flows:

- Airtime purchase.
- Bundle offers lookup.
- Bundle purchase/allocation.
- Partner-based API authentication using `X-App-Name` and `X-Api-Key`.
- Transaction persistence and idempotency support.

More telecom services are planned and contributions are welcome.

---

## Stack

- Go
- Fiber
- PostgreSQL
- Docker
- Docker Compose

---

## Project structure

```text
taas/
├── cmd/
│   └── taas-api/
│       └── main.go
├── internal/
│   ├── bootstrap/
│   ├── config/
│   ├── db/
│   ├── http/
│   ├── middleware/
│   ├── web/
│   ├── domain/
│   │   ├── partners/
│   │   ├── credentials/
│   │   └── transactions/
│   ├── countries/
│   │   └── drc/
│   │       └── vodacom/
│   ├── providers/
│   │   └── drc/
│   │       └── vodacom/
│   │           ├── airtime/
│   │           ├── bundle/
│   │           ├── client/
│   │           └── shared/
│   └── util/
├── migrations/
│   └── 001_init.sql
├── .env.example
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## Quickstart

Clone the repository, build the app, start the containers, and test the health endpoints.

```bash
git clone https://github.com/abubakar508/taas.git
cd taas

go clean -cache
go mod tidy
go build ./...

docker compose build --no-cache
docker compose up
```

Then test:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Expected responses:

```json
{"status":"ok"}
```

```json
{"status":"ready"}
```

---

## Prerequisites

Before running the project, make sure you have:

- Go installed locally.
- Docker installed.
- Docker Compose plugin installed.
- Git installed.
- Access to a PostgreSQL instance if running outside Docker.

Check your versions:

```bash
go version
docker --version
docker compose version
git --version
```

---



## Clone the project

```bash
git clone https://github.com/abubakar508/taas.git
cd taas
```

---

## Run locally without Docker

### 1. Create your environment file

```bash
cp .env.example .env
```

Example `.env`:

```env
APP_NAME=taas-api
APP_ENV=development
PORT=8080
DATABASE_URL=postgres://taas:taas@localhost:5432/taas?sslmode=disable
HTTP_READ_TIMEOUT=15s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
SHUTDOWN_TIMEOUT=20s
LOG_LEVEL=info
```

### 2. Create the database

Create the database and user in PostgreSQL:

```sql
CREATE DATABASE taas;
CREATE USER taas WITH PASSWORD 'taas';
GRANT ALL PRIVILEGES ON DATABASE taas TO taas;
```

### 3. Run migrations

```bash
psql "postgres://taas:taas@localhost:5432/taas?sslmode=disable" -f migrations/001_init.sql
```

### 4. Install modules and run

```bash
go mod tidy
go run ./cmd/taas-api
```

---

## Run with Docker

Docker is the easiest and most repeatable way to run the project locally.

### 1. Build and start

```bash
docker compose up --build
```

Or in detached mode:

```bash
docker compose up --build -d
```

### 2. Check container status

```bash
docker compose ps
```

### 3. View logs

```bash
docker compose logs -f
```

API only:

```bash
docker compose logs -f api
```

Postgres only:

```bash
docker compose logs -f postgres
```

### 4. Stop containers

```bash
docker compose down
```

To remove volumes too:

```bash
docker compose down -v
```

---

## Docker Compose note

If Docker warns that the `version` field in `docker-compose.yml` is obsolete, you can safely remove the top-level `version:` line because modern Docker Compose ignores it and uses the latest supported schema instead.

---

## Health checks

TaaS exposes health endpoints for local testing and orchestration:

### Liveness

```bash
curl http://localhost:8080/healthz
```

Expected:

```json
{"status":"ok"}
```

### Readiness

```bash
curl http://localhost:8080/readyz
```

Expected:

```json
{"status":"ready"}
```

The readiness endpoint is useful to confirm the API can reach PostgreSQL and is fully ready to serve traffic.

---

## Authentication

Protected endpoints require these headers:

```http
X-App-Name: your-app-name
X-Api-Key: your-secret-api-key
```

The service hashes the API key and validates it against the active partner record stored in PostgreSQL.

---

## Seed a demo partner

Generate a SHA-256 hash for your API key:

```bash
echo -n "SUPER_SECRET_KEY" | sha256sum
```

Then insert a partner row:

```sql
INSERT INTO partners (app_name, key_hash, is_active)
VALUES ('fishy', '<SHA256_HASH_HERE>', true);
```

---

## Seed provider credentials

Insert provider credentials for Vodacom DRC:

```sql
INSERT INTO partner_provider_credentials (partner_id, country, provider, meta, is_active)
VALUES (
  '<PARTNER_UUID>',
  'drc',
  'vodacom',
  '{
    "base_url": "https://10.0.0.25:5443",
    "basic_token": "YOUR_BASE64_BASIC_TOKEN",
    "offer_username": "your_offer_username",
    "offer_password": "your_offer_password",
    "airtime_soap_url": "https://10.0.0.30:30002/payment/services/SYNCAPIRequestMgrService",
    "third_party_id": "YOUR_THIRD_PARTY_ID",
    "third_party_password": "YOUR_THIRD_PARTY_PASSWORD",
    "initiator_identifier": "YOUR_INITIATOR_IDENTIFIER",
    "security_credential": "YOUR_SECURITY_CREDENTIAL",
    "short_code": "YOUR_SHORT_CODE",
    "caller_type": "2",
    "key_owner": "1",
    "command_id": "InitTrans_2108",
    "insecure_skip_tls_verify": false
  }'::jsonb,
  true
);
```

### Provider endpoint note

The IPs and paths above are **example placeholders only**. Replace them with the exact provider-issued hostnames, IP addresses, ports, and paths shared with you privately by the telecom/provider team.

If your provider gave you a real private IP or production hostname, use that exact value in `base_url` and `airtime_soap_url` instead of the example values shown above.

---

## Current Vodacom DRC endpoints in TaaS

These are the public TaaS endpoints currently exposed by this service:

### Airtime purchase

```http
POST /api/drc/vodacom/airtime/purchase
```

### Bundle offers

```http
POST /api/drc/vodacom/bundles/offers
```

### Bundle purchase

```http
POST /api/drc/vodacom/bundles/purchase
```

### Transaction lookup

```http
GET /api/drc/vodacom/transactions/:id
```

---

## Example: bundle offers via TAAS

```bash
curl -X POST http://localhost:8080/api/drc/vodacom/bundles/offers \
  -H "X-App-Name: fishy" \
  -H "X-Api-Key: SUPER_SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "msisdn": "243820435281"
  }'
```

Example response:

```json
{
  "status": "success",
  "message": "bundle offers fetched successfully",
  "provider_session_id": "SESSION_ID_FROM_PROVIDER",
  "provider_transaction_id": "TRANSACTION_ID_FROM_PROVIDER",
  "event_detail": "Operation completed",
  "event_description": "Offers fetched successfully",
  "offers": [
    {
      "key": "bundle-1",
      "bundle_desc": "Daily data bundle",
      "amount_unconverted": 500,
      "bundle_id": "20"
    }
  ]
}
```

---

## Example: bundle purchase via TAAS

Use the `provider_session_id` and `provider_transaction_id` returned from the offers request.

```bash
curl -X POST http://localhost:8080/api/drc/vodacom/bundles/purchase \
  -H "X-App-Name: fishy" \
  -H "X-Api-Key: SUPER_SECRET_KEY" \
  -H "Idempotency-Key: 4cba4cd6-a2b1-4d4f-9747-117599b09265" \
  -H "Content-Type: application/json" \
  -d '{
    "msisdn": "243814445069",
    "bundle_id": "20",
    "amount": 500,
    "currency": "CDF",
    "provider_session_id": "SESSION_ID_FROM_OFFERS",
    "provider_transaction_id": "TRANSACTION_ID_FROM_OFFERS",
    "client_reference": "bundle-purchase-001"
  }'
```

Example response:

```json
{
  "status": "success",
  "message": "bundle purchased successfully",
  "provider_ref": "PROVIDER_REF",
  "internal_transaction_id": "INTERNAL_TRANSACTION_UUID",
  "provider_transaction_id": "PROVIDER_REF",
  "provider_status_code": "0",
  "event_description": "Request processed successfully",
  "idempotency_key": "4cba4cd6-a2b1-4d4f-9747-117599b09265"
}
```

---

## Example: airtime purchase via TAAS

```bash
curl -X POST http://localhost:8080/api/drc/vodacom/airtime/purchase \
  -H "X-App-Name: fishy" \
  -H "X-Api-Key: SUPER_SECRET_KEY" \
  -H "Idempotency-Key: 6c1b6251-6206-4b68-a08c-278abffc57d9" \
  -H "Content-Type: application/json" \
  -d '{
    "msisdn": "243823972907",
    "amount": 1000,
    "currency": "CDF",
    "client_reference": "airtime-purchase-001"
  }'
```

Example response:

```json
{
  "status": "success",
  "message": "airtime purchase accepted",
  "transaction_id": "INTERNAL_TRANSACTION_UUID",
  "idempotency_key": "6c1b6251-6206-4b68-a08c-278abffc57d9"
}
```

---

## Bundle flow summary

The current Vodacom DRC bundle flow inside TAAS works like this:

1. Generate bearer token.
2. Fetch bundle offers for the customer MSISDN.
3. Use the returned session ID and transaction ID to allocate/purchase the selected bundle.

This mirrors the provider flow used by the telecom service integration.

---

## Airtime flow summary

The current Vodacom DRC airtime flow inside TAAS uses the configured SOAP provider integration and persists the raw provider response for audit and troubleshooting.

---

## Troubleshooting

### Build error: Go version mismatch

If Docker build fails with a Go version error, update the builder image in `Dockerfile` to match your required Go version.

Example:

```dockerfile
FROM golang:1.26-alpine AS builder
```

### Duplicate declarations in Go files

If you see errors such as:

```text
redeclared in this block
```

check for duplicate files in the same package, especially under `internal/web`.

### Compose warning about version

If you see a warning that `version` in `docker-compose.yml` is obsolete, remove that top-level field from the file.

### Postgres is not ready

Check logs:

```bash
docker compose logs -f postgres
docker compose logs -f api
```

Using `depends_on` together with a healthcheck is the preferred pattern for waiting until the database is actually ready.

---

## Open-source contributions

This project is open for community contributions.

Current implemented area:

- Vodacom DRC airtime purchase.
- Vodacom DRC bundle offers.
- Vodacom DRC bundle purchase.
- Multi-tenant partner authentication.
- Transaction tracking and idempotency.

Areas open for contribution:

- More countries.
- More telecom providers.
- Better observability and metrics.
- Webhooks and callbacks.
- Automated integration tests.
- Additional telecom services such as reversals, balances, status checks, and payout flows.

Please open issues or pull requests with clear implementation notes and testing steps.

---

## License

Add your preferred open-source license in a `LICENSE` file, or update this section when the license is finalized.
