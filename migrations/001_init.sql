CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name TEXT NOT NULL UNIQUE,
    key_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS partner_provider_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    country TEXT NOT NULL,
    provider TEXT NOT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(partner_id, country, provider)
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    country TEXT NOT NULL,
    provider TEXT NOT NULL,
    type TEXT NOT NULL,
    msisdn TEXT NOT NULL,
    bundle_id TEXT NULL,
    amount NUMERIC(18,2) NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    provider_ref TEXT NULL,
    idempotency_key TEXT NULL,
    client_reference TEXT NULL,
    error_message TEXT NULL,
    provider_raw JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_partner_id ON transactions(partner_id);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency_key ON transactions(partner_id, idempotency_key);
CREATE INDEX IF NOT EXISTS idx_credentials_partner_country_provider ON partner_provider_credentials(partner_id, country, provider);
