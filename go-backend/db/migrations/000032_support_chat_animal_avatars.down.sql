ALTER TABLE support_conversations
  DROP CONSTRAINT IF EXISTS chk_support_conversations_avatar;

ALTER TABLE support_conversations
  ADD CONSTRAINT chk_support_conversations_avatar CHECK (customer_avatar IN ('female', 'male'));
