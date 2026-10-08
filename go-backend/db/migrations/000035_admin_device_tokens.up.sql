CREATE TABLE IF NOT EXISTS admin_device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(20) NOT NULL,
    device_token TEXT NOT NULL UNIQUE,
    app_version VARCHAR(30) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_device_tokens_active
    ON admin_device_tokens (is_active) WHERE is_active = TRUE;
