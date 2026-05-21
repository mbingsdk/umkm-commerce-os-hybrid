ALTER TABLE discovery_featured_items
    ADD COLUMN IF NOT EXISTS tenant_id UUID,
    ADD COLUMN IF NOT EXISTS store_id UUID,
    ADD COLUMN IF NOT EXISTS product_id UUID,
    ADD COLUMN IF NOT EXISTS placement TEXT NOT NULL DEFAULT 'home',
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

UPDATE discovery_featured_items d
SET tenant_id = s.tenant_id,
    store_id = s.id,
    product_id = NULL,
    placement = COALESCE(NULLIF(d.placement, ''), 'home')
FROM stores s
WHERE d.item_type = 'store'
  AND d.item_id = s.id
  AND d.deleted_at IS NULL
  AND (d.tenant_id IS NULL OR d.store_id IS NULL);

UPDATE discovery_featured_items d
SET tenant_id = p.tenant_id,
    store_id = p.store_id,
    product_id = p.id,
    placement = COALESCE(NULLIF(d.placement, ''), 'home')
FROM products p
WHERE d.item_type = 'product'
  AND d.item_id = p.id
  AND d.deleted_at IS NULL
  AND (d.tenant_id IS NULL OR d.product_id IS NULL);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'discovery_featured_items_tenant_fk'
    ) THEN
        ALTER TABLE discovery_featured_items
            ADD CONSTRAINT discovery_featured_items_tenant_fk
            FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'discovery_featured_items_store_fk'
    ) THEN
        ALTER TABLE discovery_featured_items
            ADD CONSTRAINT discovery_featured_items_store_fk
            FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'discovery_featured_items_product_fk'
    ) THEN
        ALTER TABLE discovery_featured_items
            ADD CONSTRAINT discovery_featured_items_product_fk
            FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'discovery_featured_items_placement_check'
    ) THEN
        ALTER TABLE discovery_featured_items
            ADD CONSTRAINT discovery_featured_items_placement_check
            CHECK (placement IN ('home', 'stores', 'products', 'category', 'city'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'discovery_featured_items_target_check'
    ) THEN
        ALTER TABLE discovery_featured_items
            ADD CONSTRAINT discovery_featured_items_target_check
            CHECK (
                (item_type = 'store' AND store_id IS NOT NULL AND product_id IS NULL)
                OR (item_type = 'product' AND product_id IS NOT NULL)
                OR item_type NOT IN ('store', 'product')
            );
    END IF;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_discovery_featured_items_admin_list
    ON discovery_featured_items (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_discovery_featured_items_public_placement
    ON discovery_featured_items (placement, item_type, sort_order, created_at DESC)
    WHERE is_active = true AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_discovery_featured_items_tenant
    ON discovery_featured_items (tenant_id)
    WHERE deleted_at IS NULL;
