CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_confirmation_id UUID,
    method TEXT NOT NULL DEFAULT 'manual_transfer'
        CHECK (method IN ('manual_transfer', 'cash', 'qris_manual', 'bank_transfer', 'cod', 'other')),
    status TEXT NOT NULL DEFAULT 'paid'
        CHECK (status IN ('waiting_confirmation', 'paid', 'failed', 'refunded')),
    amount BIGINT NOT NULL CHECK (amount >= 0),
    payer_name TEXT,
    bank_name TEXT,
    proof_url TEXT,
    note TEXT,
    paid_at TIMESTAMPTZ,
    confirmed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payment_confirmations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payer_name TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    transfer_amount BIGINT NOT NULL CHECK (transfer_amount >= 0),
    transfer_date TIMESTAMPTZ NOT NULL,
    proof_url TEXT,
    note TEXT,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'rejected')),
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    review_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'payments_payment_confirmation_fk'
    ) THEN
        ALTER TABLE payments
        ADD CONSTRAINT payments_payment_confirmation_fk
        FOREIGN KEY (payment_confirmation_id)
        REFERENCES payment_confirmations(id)
        ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_payments_tenant_order
    ON payments (tenant_id, order_id);

CREATE INDEX IF NOT EXISTS idx_payments_status
    ON payments (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_payments_created_at
    ON payments (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_confirmations_tenant_order
    ON payment_confirmations (tenant_id, order_id);

CREATE INDEX IF NOT EXISTS idx_payment_confirmations_status
    ON payment_confirmations (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_payment_confirmations_created_at
    ON payment_confirmations (tenant_id, created_at DESC);
