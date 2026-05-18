CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    price_monthly BIGINT NOT NULL DEFAULT 0,
    product_limit INT NOT NULL DEFAULT 100,
    staff_limit INT NOT NULL DEFAULT 2,
    can_use_pos BOOLEAN NOT NULL DEFAULT true,
    can_use_discovery BOOLEAN NOT NULL DEFAULT true,
    can_use_courier BOOLEAN NOT NULL DEFAULT false,
    can_use_custom_domain BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (price_monthly >= 0),
    CHECK (product_limit >= 0),
    CHECK (staff_limit >= 0)
);

INSERT INTO plans (code, name)
VALUES ('starter', 'Starter')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID REFERENCES plans(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'trialing'
        CHECK (status IN ('active', 'trialing', 'suspended', 'cancelled')),
    trial_ends_at TIMESTAMPTZ,
    subscription_ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS user_tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role TEXT NOT NULL
        CHECK (role IN ('owner', 'manager', 'staff', 'cashier', 'inventory_staff', 'courier_admin', 'driver')),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'invited', 'disabled')),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    joined_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, tenant_id)
);

CREATE TABLE IF NOT EXISTS stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    logo_url TEXT,
    banner_url TEXT,
    phone TEXT,
    whatsapp TEXT,
    email TEXT,
    address TEXT,
    city TEXT,
    province TEXT,
    postal_code TEXT,
    latitude NUMERIC(10, 7),
    longitude NUMERIC(10, 7),
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'unpublished', 'suspended')),
    is_discoverable BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS store_business_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),
    open_time TIME,
    close_time TIME,
    is_closed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (store_id, day_of_week),
    CHECK (
        is_closed = true
        OR (open_time IS NOT NULL AND close_time IS NOT NULL AND open_time < close_time)
    )
);

CREATE TABLE IF NOT EXISTS tenant_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID REFERENCES stores(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id UUID,
    before_data JSONB,
    after_data JSONB,
    reason TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_tenants_user_id
    ON user_tenants (user_id);

CREATE INDEX IF NOT EXISTS idx_user_tenants_tenant_id
    ON user_tenants (tenant_id);

CREATE INDEX IF NOT EXISTS idx_stores_tenant_id
    ON stores (tenant_id);

CREATE INDEX IF NOT EXISTS idx_store_business_hours_tenant_id_store_id
    ON store_business_hours (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_tenant_id_created_at
    ON tenant_audit_logs (tenant_id, created_at DESC);
