DROP INDEX IF EXISTS idx_support_conversations_inbox;
ALTER TABLE support_conversations
  DROP COLUMN IF EXISTS last_sender,
  DROP COLUMN IF EXISTS last_message_preview,
  DROP COLUMN IF EXISTS admin_unread_count;
