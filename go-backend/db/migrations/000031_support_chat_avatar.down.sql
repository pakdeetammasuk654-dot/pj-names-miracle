ALTER TABLE support_conversations
  DROP CONSTRAINT IF EXISTS chk_support_conversations_avatar;

ALTER TABLE support_conversations
  DROP COLUMN IF EXISTS customer_avatar;
