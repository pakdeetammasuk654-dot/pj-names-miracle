# Codex Development Guide

## Project map

- `android-frontend/`: Native Android app, Kotlin, Jetpack Compose, MVVM via `NameSearchViewModel`.
- `go-backend/`: Go/Gin API using Clean Architecture and a pipeline for name search.
- `project_context.md`: product and architecture source of truth. Keep it current when product rules change.

## Required product rules

- Build Android UI with Jetpack Compose. Do not introduce XML layouts for ordinary UI.
- Keep numerology, Taksa, ranking, semantic search, and other business calculations in Go. Android should own UI state and rendering only.
- Use the Thai terminology `เลขศาสตร์` and `พลังเงา`.
- Render every number pair with the shared colored-circle presentation (`PairCircles`), never as plain bracket text.

## Local development

- Backend: `cd go-backend && go run main.go` (loads local `.env` automatically)
- Backend verification: `cd go-backend && go test ./...`
- Android verification: `cd android-frontend && ./gradlew :app:assembleDebug`
- Android emulator points to `http://10.0.2.2:8080/`. Make endpoint configuration explicit before supporting physical devices or release builds.

## Safety and quality

- Never commit `.env`, API keys, passwords, APKs, Gradle output, or the local Go binary.
- Do not add secret fallback values to source code. Use environment variables.
- Preserve user changes and avoid unrelated refactors.
- Add or update tests when changing backend business logic. Run the relevant verification command before handoff.
