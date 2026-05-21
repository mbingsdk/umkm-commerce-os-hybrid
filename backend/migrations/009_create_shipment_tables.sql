CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    courier_type TEXT NOT NULL DEFAULT 'manual'
        CHECK (courier_type IN ('internal', 'manual')),
    courier_name TEXT,
    tracking_number TEXT,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'ready_for_pickup', 'picked_up', 'on_delivery', 'delivered', 'failed', 'cancelled')),
    shipping_cost BIGINT NOT NULL DEFAULT 0 CHECK (shipping_cost >= 0),
    assigned_to_name TEXT,
    assigned_to_phone TEXT,
    note TEXT,
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shipment_status_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    note TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_shipments_tenant_store
    ON shipments (tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_shipments_order
    ON shipments (tenant_id, store_id, order_id);

CREATE INDEX IF NOT EXISTS idx_shipments_status
    ON shipments (tenant_id, store_id, status);

CREATE INDEX IF NOT EXISTS idx_shipments_tracking_number
    ON shipments (tenant_id, tracking_number);

CREATE INDEX IF NOT EXISTS idx_shipment_status_logs_shipment
    ON shipment_status_logs (shipment_id, created_at ASC);
