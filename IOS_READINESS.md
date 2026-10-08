# iOS readiness contract

This repository is ready to start an iOS client after every automated gate below passes. iOS must be a SwiftUI/MVVM API client; numerology, พลังเงา, Kalakini, semantic retrieval, and ranking remain exclusively in Go.

## Stable backend contract

- Base URL: `https://xn--b3cu8e7ah6h.com/api/v1`
- API header: `X-API-Version: 1`
- Ranking header and policy field: `adaptive_pairs_v1`
- Machine-readable schema: `go-backend/docs/openapi-v1.yaml`
- Before search, call `GET /readiness`. Retry a `503` with bounded exponential backoff; never interpret it as an empty ranking.
- Preserve `candidate_token` only in memory. A `404` from rerank means the pool expired; repeat the original search once.
- Render server-provided pair strings and types with the same colored-circle rules as Android. Never recompute business rules on-device.
- Use `requested_weights` for user intent and `applied_weights` for confirmed selector/results state. Swap rows and applied weights atomically.
- Ignore unknown JSON fields for forward-compatible additive changes. Fail closed if the API or ranking contract header is unsupported.

## Release gates before TestFlight

- [x] Backend business logic has Go unit/integration coverage.
- [x] Search/rerank use an opaque cached pool; rerank does not repeat DB/AI retrieval.
- [x] Adaptive pair policy is explicit and versioned.
- [x] Production exposes database-backed readiness.
- [ ] Sound-calibration approval: product owner approved proceeding on 2026-09-15, but the immutable corpus snapshot and held-out human evaluation artifact still do not exist. Provisional sound scores must not be marketed as independently validated.
- [ ] Configure deployment health checks against `/api/v1/readiness` and retain rollback binary.
- [ ] Implement Swift decoding contract tests from fixed JSON fixtures.
- [ ] Implement StoreKit 2 purchase, restore, interrupted purchase, and server-independent entitlement recovery tests.
- [x] Add the app Privacy Manifest and declare the UserDefaults required reason used for one-time migration of legacy local credentials.
- [x] Add customer-controlled chat deletion on Android, iOS, and the API. Deletion removes the conversation and cascading message/device-token records plus its uploaded chat images; the public privacy policy now describes chat, notifications, purchases, website analytics, and the uninstall limitation.
- [x] Deploy the API deletion endpoint and version-aware privacy policy to production on 2026-09-15. Verified `/api/v1/readiness` returns `READY`, an unauthenticated delete returns 401, and a disposable production conversation was created, deleted with its token (204), and confirmed inaccessible (401). Backend SHA-256: `858cc6f6f603522c79200fd1a4f1eb811ee5053375784825c93bd6b1c2faf9ed`. Rollback binary: `/opt/pj-names-miracle/backups/appstore-delete-fix-20260915T091325Z/pj-names-miracle-server`.
- [x] Publish App Privacy answers in App Store Connect on 2026-09-15: Name, Photos or Videos, Customer Support, Search History, User ID, and Device ID; each is marked for App Functionality, linked to the user's identity, and not used for tracking. Provider-specific data handling and retention still need verification.
- [ ] Complete ATS review, verify the public support/privacy URLs, and account-deletion workflow if accounts are introduced.
- [ ] Run TestFlight tests for cold launch, offline mode, token expiry, 503 retry, stale rerank rejection, Dynamic Type, VoiceOver, Thai text wrapping, and iPhone/iPad layouts.
- [x] Existing App Store version `1.0.9 (17)` is approved and publicly listed in Thailand as of 2026-09-15; verified at https://apps.apple.com/th/app/id6800746663.
- [x] Increment the iOS App Store version to `1.0.10 (20)` and locally archive/export a signed IPA at `/tmp/ChueDee-AppStore-1.0.10/exported/ChueDee.ipa`. Xcode reports archive/export success; the archive signature verifies as valid, bundle ID is `com.numberniceic.app`, and the IPA has not been uploaded.
- [x] Correct and publish the App Privacy disclosure: App Store Connect confirms “Data Linked to You” for six data types. The listing preview no longer says “Data Not Collected.”
- [ ] Upload and submit IPA `1.0.10 (20)` after completing the remaining release gates. The App Privacy disclosure, production deletion API, and privacy policy are now updated; build 20 remains local and has not been uploaded.

## Product parity rule

Android and iOS must send identical normalized inputs and display the backend's returned order. A cross-platform golden fixture should compare `name_id`, rank, pair types, risk metadata, final score, requested/applied weights, and policy versions—not screenshots or locally recomputed values.
