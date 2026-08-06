# Miracle Names

Android and Go application for auspicious Thai-name analysis. The product uses numerology, Taksa/Kalakini rules, and semantic ranking to suggest and explain names.

## Structure

| Directory | Purpose |
| --- | --- |
| `android-frontend/` | Kotlin/Jetpack Compose Android client |
| `go-backend/` | Go/Gin API, search pipeline, PostgreSQL repository |
| `project_context.md` | Product and architecture source of truth |
| `AGENTS.md` | Working rules for Codex contributors |

## Prerequisites

- Go 1.25.5
- JDK 17+ (the Android project targets JVM 17)
- Android SDK for API 34 and an Android emulator/device
- PostgreSQL with the project schema, including `pgvector` and `pg_trgm` for semantic search

## Run locally

1. Create local backend configuration:

   ```sh
   cd go-backend
   cp .env.example .env
   # Edit .env with your local development credentials.
   ```

2. Start the API:

   ```sh
   ./run-local.sh
   ```

   It listens on `http://localhost:8080`. The Android emulator uses `10.0.2.2` to reach this host.

3. Build the Android app:

   ```sh
   cd ../android-frontend
   ./gradlew :app:assembleDebug
   ```

## Verify

```sh
cd go-backend && go test ./...
cd ../android-frontend && ./gradlew :app:assembleDebug
```

## Security

Keep credentials only in the untracked `go-backend/.env` file or your process environment. If a credential was ever committed or pasted into source code, revoke and replace it before sharing the repository.

## Development rules

Read [project_context.md](project_context.md) and [AGENTS.md](AGENTS.md) before changing product behavior. In brief: Compose-only Android UI, Go owns business logic, and all pair numbers use colored circles in the UI.
