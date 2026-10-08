ALTER TABLE support_conversations
  ADD COLUMN IF NOT EXISTS admin_unread_count INTEGER NOT NULL DEFAULT 0 CHECK (admin_unread_count >= 0),
  ADD COLUMN IF NOT EXISTS last_message_preview TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_sender TEXT NOT NULL DEFAULT '' CHECK (last_sender IN ('', 'customer', 'admin'));

CREATE INDEX IF NOT EXISTS idx_support_conversations_inbox
  ON support_conversations (status, admin_unread_count DESC, updated_at DESC);
