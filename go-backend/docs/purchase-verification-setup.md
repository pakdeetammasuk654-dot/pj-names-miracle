# Production purchase verification setup

The `premium_lifetime` name-suggestion entitlement is issued only after the Go backend checks the purchase with the relevant store. Configure credentials on the production server, not in the app or Git. A missing credential must return a temporary verification error, never an entitlement.

## Google Play

1. In Google Cloud, create a dedicated service account for this app and enable the Google Play Developer API in its project.
2. In Google Play Console, open **Users and permissions**, invite that service account's email, and grant the permissions Google requires for Billing API access: **View financial data, orders, and cancellation survey responses** and **Manage orders and subscriptions**, scoped to `com.numberniceic.app` where the console permits.
3. Create and download a JSON key for that service account. Keep the only server copy outside the repository, for example `/etc/pj-names-miracle/credentials/google-play-purchases.json`, readable by the backend service account only.
4. Set `GOOGLE_PLAY_SERVICE_ACCOUNT_PATH` to that absolute path and `GOOGLE_PLAY_PACKAGE_NAME=com.numberniceic.app` in `/etc/pj-names-miracle/pj-names-miracle.env`.

Do not reuse a Firebase messaging service account merely because it is already present; a dedicated account limits the effect of a leaked key.

## App Store

1. In App Store Connect, open **Users and Access → Integrations → In-App Purchase** and generate an **In-App Purchase** key for App Store Server API. This is not an APNs key or a generic App Store Connect API key.
2. Record the key ID and issuer ID, and download the `.p8` private key immediately: Apple offers the private-key download only once.
3. Keep the server copy outside the repository, for example `/etc/pj-names-miracle/credentials/app-store-in-app-purchase.p8`, readable by the backend service account only.
4. Set `APP_STORE_ISSUER_ID`, `APP_STORE_KEY_ID`, `APP_STORE_PRIVATE_KEY_PATH` (absolute path), and `APP_STORE_BUNDLE_ID=com.numberniceic.app` in `/etc/pj-names-miracle/pj-names-miracle.env`.

## Release checks

- Keep the existing `ENTITLEMENT_SIGNING_SECRET` unchanged so already-issued ranking tokens remain valid. It must be at least 32 bytes and must never be printed, committed, or put into a client app.
- Restart the backend only after both files and environment entries exist. Check file ownership/mode and service logs without displaying credential contents.
- Verify a genuine Google Play test purchase and an Apple Sandbox purchase through the respective app, then restore each purchase. An invented purchase token or transaction ID must fail.
- Confirm `GET /api/v1/names/suggestions` returns only count/pair previews without an entitlement and full names only with a store-verified entitlement.
- iOS Simulator StoreKit local testing and Android debug mock mode verify UI presentation only; they cannot prove production server purchase verification.
