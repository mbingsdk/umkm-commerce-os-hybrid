ALTER TABLE plans
    ADD COLUMN IF NOT EXISTS description TEXT;

ALTER TABLE plans
    ALTER COLUMN product_limit DROP NOT NULL,
    ALTER COLUMN staff_limit DROP NOT NULL;

ALTER TABLE plans
    DROP CONSTRAINT IF EXISTS plans_product_limit_check,
    DROP CONSTRAINT IF EXISTS plans_staff_limit_check,
    DROP CONSTRAINT IF EXISTS plans_product_limit_non_negative,
    DROP CONSTRAINT IF EXISTS plans_staff_limit_non_negative;

ALTER TABLE plans
    ADD CONSTRAINT plans_product_limit_non_negative
        CHECK (product_limit IS NULL OR product_limit >= 0),
    ADD CONSTRAINT plans_staff_limit_non_negative
        CHECK (staff_limit IS NULL OR staff_limit >= 0);

ALTER TABLE outbox_events
    ALTER COLUMN tenant_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_plans_active_code
    ON plans (is_active, code);
