ALTER TABLE users
    ADD COLUMN IF NOT EXISTS platform_role TEXT NOT NULL DEFAULT 'user';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_platform_role_check'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_platform_role_check
            CHECK (platform_role IN ('user', 'super_admin'));
    END IF;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_users_platform_role
    ON users (platform_role);

CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    target_type TEXT,
    target_id UUID,
    before_data JSONB,
    after_data JSONB,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (action <> '')
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_actor
    ON admin_audit_logs (actor_user_id);

CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_target
    ON admin_audit_logs (target_type, target_id);

CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_created_at
    ON admin_audit_logs (created_at DESC);
