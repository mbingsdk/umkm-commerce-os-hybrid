CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT,
    notes TEXT,
    total_orders INT NOT NULL DEFAULT 0,
    total_spent BIGINT NOT NULL DEFAULT 0,
    last_order_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, phone),
    CHECK (total_orders >= 0),
    CHECK (total_spent >= 0)
);

CREATE TABLE IF NOT EXISTS customer_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    label TEXT,
    recipient_name TEXT NOT NULL,
    recipient_phone TEXT NOT NULL,
    address TEXT NOT NULL,
    city TEXT,
    province TEXT,
    postal_code TEXT,
    latitude NUMERIC(10, 7),
    longitude NUMERIC(10, 7),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    order_number TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'storefront'
        CHECK (source IN ('storefront', 'marketplace_discovery', 'pos', 'whatsapp_manual', 'admin_manual', 'marketplace_sync', 'reseller', 'api_partner')),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'processing', 'ready_to_ship', 'shipped', 'delivered', 'completed', 'cancelled', 'returned', 'refunded')),
    payment_status TEXT NOT NULL DEFAULT 'unpaid'
        CHECK (payment_status IN ('unpaid', 'waiting_confirmation', 'paid', 'failed', 'refunded')),
    shipment_status TEXT
        CHECK (shipment_status IS NULL OR shipment_status IN ('pending', 'ready_for_pickup', 'picked_up', 'on_delivery', 'delivered', 'failed', 'returned', 'cancelled')),
    subtotal BIGINT NOT NULL DEFAULT 0,
    discount_total BIGINT NOT NULL DEFAULT 0,
    shipping_cost BIGINT NOT NULL DEFAULT 0,
    tax_total BIGINT NOT NULL DEFAULT 0,
    grand_total BIGINT NOT NULL DEFAULT 0,
    customer_name TEXT NOT NULL,
    customer_phone TEXT NOT NULL,
    customer_email TEXT,
    shipping_address TEXT,
    shipping_city TEXT,
    shipping_province TEXT,
    shipping_postal_code TEXT,
    customer_note TEXT,
    internal_note TEXT,
    confirmed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, order_number),
    CHECK (subtotal >= 0),
    CHECK (discount_total >= 0),
    CHECK (shipping_cost >= 0),
    CHECK (tax_total >= 0),
    CHECK (grand_total >= 0)
);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name TEXT NOT NULL,
    sku TEXT,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price BIGINT NOT NULL CHECK (unit_price >= 0),
    discount_total BIGINT NOT NULL DEFAULT 0 CHECK (discount_total >= 0),
    subtotal BIGINT NOT NULL CHECK (subtotal >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_status_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    note TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS stock_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity > 0),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'confirmed', 'released', 'expired', 'cancelled')),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    released_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_body JSONB,
    status_code INT,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, scope, idempotency_key),
    CHECK (scope <> ''),
    CHECK (idempotency_key <> ''),
    CHECK (request_hash <> ''),
    CHECK (status_code IS NULL OR status_code BETWEEN 100 AND 599)
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'processed', 'failed')),
    attempts INT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (event_type <> ''),
    CHECK (aggregate_type <> '')
);

CREATE INDEX IF NOT EXISTS idx_customers_tenant_store
    ON customers (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_customers_phone
    ON customers (phone);

CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer
    ON customer_addresses (customer_id);

CREATE INDEX IF NOT EXISTS idx_orders_tenant_store
    ON orders (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id
    ON orders (customer_id);

CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_orders_payment_status
    ON orders (tenant_id, payment_status);

CREATE INDEX IF NOT EXISTS idx_orders_created_at
    ON orders (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_customer_phone
    ON orders (tenant_id, customer_phone);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id
    ON order_items (order_id);

CREATE INDEX IF NOT EXISTS idx_order_items_product_id
    ON order_items (product_id);

CREATE INDEX IF NOT EXISTS idx_order_status_logs_order_id
    ON order_status_logs (order_id, created_at ASC);

CREATE INDEX IF NOT EXISTS idx_stock_reservations_tenant_order
    ON stock_reservations (tenant_id, order_id);

CREATE INDEX IF NOT EXISTS idx_stock_reservations_status
    ON stock_reservations (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_stock_reservations_expires_at
    ON stock_reservations (expires_at);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_tenant_scope
    ON idempotency_keys (tenant_id, scope);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_locked_until
    ON idempotency_keys (locked_until);

CREATE INDEX IF NOT EXISTS idx_outbox_events_status_available
    ON outbox_events (status, available_at);

CREATE INDEX IF NOT EXISTS idx_outbox_events_tenant
    ON outbox_events (tenant_id);

CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate
    ON outbox_events (aggregate_type, aggregate_id);

CREATE INDEX IF NOT EXISTS idx_outbox_events_created_at
    ON outbox_events (created_at ASC);
