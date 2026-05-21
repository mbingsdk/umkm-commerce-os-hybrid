CREATE TABLE IF NOT EXISTS discovery_featured_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_type TEXT NOT NULL
        CHECK (item_type IN ('store', 'product', 'category', 'banner')),
    item_id UUID,
    title TEXT,
    description TEXT,
    image_url TEXT,
    target_url TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS idx_discovery_featured_items_active
    ON discovery_featured_items (item_type, sort_order, created_at DESC)
    WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_discovery_featured_items_item
    ON discovery_featured_items (item_type, item_id);
