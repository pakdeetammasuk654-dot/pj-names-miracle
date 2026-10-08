# Codex Development Guide

## Project map

- `android-frontend/`: Native Android app, Kotlin, Jetpack Compose, MVVM via `NameSearchViewModel`.
- `ios-frontend/`: Native iOS app, SwiftUI, MVVM via `SearchViewModel`.
- `go-backend/`: Go/Gin API using Clean Architecture and a pipeline for name search.
- `project_context.md`: product and architecture source of truth. Keep it current when product rules change.

## Required product rules

- Application ID / Package Name MUST be `com.numberniceic.app` (App Title: "ชื่อดี 4 มิติ", Internal: "ชื่อดี 4 มิติ").
- Build Android UI with Jetpack Compose. Do not introduce XML layouts for ordinary UI.
- Keep numerology, Kalakini, ranking, semantic search, and other business calculations in Go. Android should own UI state and rendering only.
- Use the Thai terminology `เลขศาสตร์` and `พลังเงา`.
- Render every number pair with the shared colored-circle presentation (`PairCircles`), never as plain bracket text.
- When a user-facing feature or UI changes, implement and verify the equivalent change in both `android-frontend/` and `ios-frontend/` unless the user explicitly scopes it to one platform.

## Local development

- Backend: `cd go-backend && go run main.go` (loads local `.env` automatically)
- Backend verification: `cd go-backend && go test ./...`
- Workflow Rule: Always develop and test Go backend changes locally first, then deploy to production server `43.228.85.200` to verify integration with the Android frontend.
- Deployment Rule: "Deploy" strictly means releasing/submitting directly to **Production** (live for real end-users on Google Play Production track & App Store Production), never internal/testing tracks unless explicitly requested by the user.
- Android verification: `cd android-frontend && ./gradlew :app:assembleDebug`
- iOS verification: `cd ios-frontend && xcodebuild -project ChueDee.xcodeproj -scheme ChueDee -sdk iphoneos -configuration Debug CODE_SIGNING_ALLOWED=NO build`
- Android emulator points to `http://10.0.2.2:8080/`. Make endpoint configuration explicit before supporting physical devices or release builds.

## Safety and quality

- Never commit `.env`, API keys, passwords, APKs, Gradle output, or the local Go binary.
- Do not add secret fallback values to source code. Use environment variables.
- Preserve user changes and avoid unrelated refactors.
- Add or update tests when changing backend business logic. Run the relevant verification command before handoff.
