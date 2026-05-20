CREATE TABLE IF NOT EXISTS expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID REFERENCES stores(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CHECK (name <> ''),
    CHECK (slug <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_expense_categories_system_slug
    ON expense_categories (slug)
    WHERE tenant_id IS NULL
      AND store_id IS NULL
      AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_expense_categories_tenant_store_slug
    ON expense_categories (tenant_id, store_id, slug)
    WHERE tenant_id IS NOT NULL
      AND store_id IS NOT NULL
      AND deleted_at IS NULL;

INSERT INTO expense_categories (name, slug, is_system)
VALUES
    ('Operasional', 'operasional', true),
    ('Bahan Baku', 'bahan_baku', true),
    ('Gaji', 'gaji', true),
    ('Pengiriman', 'pengiriman', true),
    ('Marketing', 'marketing', true),
    ('Lainnya', 'lainnya', true)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id UUID REFERENCES expense_categories(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    expense_date DATE NOT NULL,
    payment_method TEXT
        CHECK (payment_method IS NULL OR payment_method IN ('cash', 'bank_transfer', 'qris_manual', 'other')),
    note TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CHECK (title <> '')
);

CREATE INDEX IF NOT EXISTS idx_expense_categories_tenant_store
    ON expense_categories (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_expenses_tenant_store_date
    ON expenses (tenant_id, store_id, expense_date DESC, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_category
    ON expenses (tenant_id, category_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_created_by
    ON expenses (created_by);
