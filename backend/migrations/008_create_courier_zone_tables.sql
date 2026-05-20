CREATE TABLE IF NOT EXISTS courier_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    rate BIGINT NOT NULL DEFAULT 0 CHECK (rate >= 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_courier_zones_tenant_store
    ON courier_zones (tenant_id, store_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_courier_zones_public_active
    ON courier_zones (tenant_id, store_id, sort_order, name)
    WHERE is_active = true AND deleted_at IS NULL;
