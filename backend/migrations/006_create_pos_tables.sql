CREATE TABLE IF NOT EXISTS cashier_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    session_number TEXT NOT NULL,
    opening_cash BIGINT NOT NULL DEFAULT 0 CHECK (opening_cash >= 0),
    closing_cash BIGINT CHECK (closing_cash IS NULL OR closing_cash >= 0),
    expected_cash BIGINT CHECK (expected_cash IS NULL OR expected_cash >= 0),
    difference BIGINT,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closed', 'cancelled')),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, session_number)
);

CREATE TABLE IF NOT EXISTS pos_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_session_id UUID NOT NULL REFERENCES cashier_sessions(id) ON DELETE RESTRICT,
    cashier_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    transaction_number TEXT NOT NULL,
    subtotal BIGINT NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    discount_total BIGINT NOT NULL DEFAULT 0 CHECK (discount_total >= 0),
    tax_total BIGINT NOT NULL DEFAULT 0 CHECK (tax_total >= 0),
    grand_total BIGINT NOT NULL DEFAULT 0 CHECK (grand_total >= 0),
    payment_method TEXT NOT NULL
        CHECK (payment_method IN ('cash', 'bank_transfer', 'qris_manual', 'other')),
    payment_amount BIGINT NOT NULL DEFAULT 0 CHECK (payment_amount >= 0),
    change_amount BIGINT NOT NULL DEFAULT 0 CHECK (change_amount >= 0),
    status TEXT NOT NULL DEFAULT 'completed'
        CHECK (status IN ('completed', 'cancelled', 'refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, transaction_number)
);

CREATE TABLE IF NOT EXISTS pos_transaction_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    pos_transaction_id UUID NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name TEXT NOT NULL,
    sku TEXT,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price BIGINT NOT NULL CHECK (unit_price >= 0),
    discount_total BIGINT NOT NULL DEFAULT 0 CHECK (discount_total >= 0),
    subtotal BIGINT NOT NULL CHECK (subtotal >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cashier_sessions_tenant_store
    ON cashier_sessions (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_cashier_sessions_cashier
    ON cashier_sessions (cashier_id);

CREATE INDEX IF NOT EXISTS idx_cashier_sessions_status
    ON cashier_sessions (tenant_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cashier_one_open_session
    ON cashier_sessions (tenant_id, store_id, cashier_id)
    WHERE status = 'open';

CREATE INDEX IF NOT EXISTS idx_pos_transactions_tenant_store
    ON pos_transactions (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_pos_transactions_session
    ON pos_transactions (cashier_session_id);

CREATE INDEX IF NOT EXISTS idx_pos_transactions_created_at
    ON pos_transactions (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_pos_transaction_items_transaction
    ON pos_transaction_items (pos_transaction_id);

CREATE INDEX IF NOT EXISTS idx_pos_transaction_items_product
    ON pos_transaction_items (product_id);
