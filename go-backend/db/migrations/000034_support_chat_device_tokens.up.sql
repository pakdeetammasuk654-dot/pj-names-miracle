CREATE TABLE IF NOT EXISTS support_chat_device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES support_conversations(id) ON DELETE CASCADE,
    platform VARCHAR(20) NOT NULL,
    device_token TEXT NOT NULL,
    app_version VARCHAR(30) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_support_chat_device_tokens UNIQUE (conversation_id, device_token)
);

CREATE INDEX IF NOT EXISTS idx_support_chat_device_tokens_conv
    ON support_chat_device_tokens (conversation_id) WHERE is_active = TRUE;
