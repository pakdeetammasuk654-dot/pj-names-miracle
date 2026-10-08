ALTER TABLE support_conversations
  ALTER COLUMN customer_avatar SET DEFAULT 'cat';

ALTER TABLE support_conversations
  DROP CONSTRAINT IF EXISTS chk_support_conversations_avatar;

ALTER TABLE support_conversations
  ADD CONSTRAINT chk_support_conversations_avatar CHECK (customer_avatar IN ('cat', 'dog', 'rabbit', 'bear', 'panda', 'female', 'male'));
