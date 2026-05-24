CREATE INDEX IF NOT EXISTS idx_products_store_public_listing
    ON products (tenant_id, store_id, status, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_store_category_created
    ON products (tenant_id, store_id, category_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_store_slug_lookup
    ON products (tenant_id, store_id, slug)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_discovery_listing
    ON products (status, is_discoverable, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_pos_search
    ON products (tenant_id, store_id, status, name, id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_images_primary_lookup
    ON product_images (tenant_id, product_id, is_primary DESC, sort_order ASC, created_at ASC);

CREATE INDEX IF NOT EXISTS idx_product_stock_snapshots_product_tenant_store
    ON product_stock_snapshots (product_id, tenant_id, store_id);

CREATE INDEX IF NOT EXISTS idx_product_stock_snapshots_low_stock
    ON product_stock_snapshots (tenant_id, store_id, quantity_available, low_stock_threshold, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_stock_movements_tenant_store_product_created
    ON stock_movements (tenant_id, store_id, product_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_orders_tenant_store_created
    ON orders (tenant_id, store_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_orders_tenant_store_status_created
    ON orders (tenant_id, store_id, status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_orders_tenant_store_payment_created
    ON orders (tenant_id, store_id, payment_status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_orders_tenant_store_source_created
    ON orders (tenant_id, store_id, source, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_orders_finance_paid_range
    ON orders (tenant_id, store_id, payment_status, status, (COALESCE(paid_at, updated_at, created_at)));

CREATE INDEX IF NOT EXISTS idx_pos_transactions_tenant_store_session_created
    ON pos_transactions (tenant_id, store_id, cashier_session_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_pos_transactions_finance_range
    ON pos_transactions (tenant_id, store_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_expenses_tenant_store_date_active
    ON expenses (tenant_id, store_id, expense_date DESC, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_events_pending_attempts_available
    ON outbox_events (status, attempts, available_at, created_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_stores_public_discovery
    ON stores (status, is_discoverable, city, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_categories_store_slug_active
    ON categories (tenant_id, store_id, slug)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tenants_status_plan_created
    ON tenants (status, plan_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_tenants_tenant_role_status_created
    ON user_tenants (tenant_id, role, status, created_at ASC);
