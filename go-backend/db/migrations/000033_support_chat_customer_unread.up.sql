ALTER TABLE support_conversations
  ADD COLUMN IF NOT EXISTS customer_unread_count INTEGER NOT NULL DEFAULT 0 CHECK (customer_unread_count >= 0);
