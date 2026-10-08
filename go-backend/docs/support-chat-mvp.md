# Support chat MVP (local verification)

This release adds an end-to-end support chat: a customer opens `/chat`, sends
text or one image, and an authenticated administrator replies from
`/admin/chat`.

## Local setup

1. Apply `db/migrations/000029_support_chat.up.sql` and
   `db/migrations/000030_support_chat_inbox.up.sql` to the local PostgreSQL
   database using the project's configured migration process.
2. Set the existing `ARTICLE_ADMIN_USERNAME` and `ARTICLE_ADMIN_PASSWORD` in
   the local environment. The chat inbox uses the same Basic Auth credentials.
3. Run `go run main.go`.
4. Open `http://localhost:8080/chat` as the customer and
   `http://localhost:8080/admin/chat` as the administrator.

The customer page stores its opaque conversation token in browser localStorage,
so the same browser can reopen the existing conversation without registration.

## API contract for native clients

- `POST /api/v1/chat/conversations` accepts `customer_name` and returns a
  `conversation_id` plus an opaque `customer_token`.
- `GET|POST /api/v1/chat/conversations/{id}/messages` require the
  `X-Chat-Token` header. A message body uses multipart form fields `body` and
  optional `image`.
- Images are checked by file signature and limited to 8 MB (JPG, PNG, WEBP).

The native app must store `customer_token` in platform-secure storage. This MVP
does not yet collect consultation data or send push notifications.

## Admin inbox

`/admin/chat` opens on **รอแอดมิน** by default. It loads a maximum of 50
conversations per view, sorted by pending customer messages and latest activity.
After an administrator replies, the conversation moves to **กำลังคุย**. A
closed conversation moves back to **รอแอดมิน** automatically when the customer
sends a new message.
