-- The first support-chat release deliberately stores only a display name and
-- an opaque client token. Personal consultation details are collected later,
-- only when a customer chooses to provide them.
CREATE TABLE IF NOT EXISTS support_conversations (
  id UUID PRIMARY KEY,
  customer_token UUID NOT NULL UNIQUE,
  customer_name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_messages (
  id UUID PRIMARY KEY,
  conversation_id UUID NOT NULL REFERENCES support_conversations(id) ON DELETE CASCADE,
  sender TEXT NOT NULL CHECK (sender IN ('customer', 'admin')),
  body TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (length(trim(body)) > 0 OR image_url <> '')
);

CREATE INDEX IF NOT EXISTS idx_support_conversations_updated_at
  ON support_conversations (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_messages_conversation_created_at
  ON support_messages (conversation_id, created_at ASC);
