ALTER TABLE support_conversations
  ADD COLUMN IF NOT EXISTS customer_avatar TEXT NOT NULL DEFAULT 'female';

ALTER TABLE support_conversations
  DROP CONSTRAINT IF EXISTS chk_support_conversations_avatar;

ALTER TABLE support_conversations
  ADD CONSTRAINT chk_support_conversations_avatar CHECK (customer_avatar IN ('female', 'male'));
