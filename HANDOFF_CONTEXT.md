# Handoff Context Document: ChueDee ("ชื่อดี 4 มิติ")

> **Last Updated**: 2026-09-26 12:22 (Bangkok Time)  
> **Target Audience**: AI Models (Gemini, Claude, GPT, etc.) and Software Engineers picking up development.

---

## 1. Project Overview & Repository Map

This repository (`pj-names-miracle`) is a production multi-platform auspicious name analysis and 4D ranking ecosystem ("ชื่อดี 4 มิติ").

### Primary Components:
- **`go-backend/`**: Go 1.22+ / Gin web framework adhering to Clean Architecture (`cmd`, `delivery`, `domain`, `repository`, `usecase`).
  - Central computing hub: handles numerology, shadow power, Kalakini filters, phonetic sound calibration, pg_trgm similarity, vector embeddings, store entitlement verification, and admin backoffices.
- **`android-frontend/`**: Native Android app written in 100% Jetpack Compose (Kotlin) with MVVM architecture.
  - Package Name: `com.numberniceic.app`
  - Current Version: v1.4.14 (versionCode 31)
- **`ios-frontend/`**: Native iOS 16+ SwiftUI app with MVVM architecture (`ChueDee.xcodeproj`).
  - Bundle Identifier: `com.numberniceic.app`
  - Current Version: 1.0.11 (Build 21)
- **`project_context.md`**: The Single Source of Truth (SSOT) for product rules and architectural definitions.

---

## 2. Critical Rules & Guardrails (`AGENTS.md`)

1. **Package / Bundle ID**: Must strictly remain `com.numberniceic.app` on both Android and iOS.
2. **No XML Layouts on Android**: All UI must be written in Jetpack Compose.
3. **Heavy Backend / Lightweight Client**: Numerology, Kalakini, phonetic scoring, ranking, and semantic search are 100% computed in Go backend. Android/iOS only manage UI state and display.
4. **Mandatory Thai Terminology**:
   - Primary Sum: `เลขศาสตร์` (SatSum)
   - Shadow/Consonant Sum: `พลังเงา` (ShaSum)
5. **Pair Presentation (`PairCircles`)**: Number pairs must ALWAYS be rendered as colored circles (`PairCircles`), never as raw bracketed text.
6. **Platform Parity**: User-facing features and fixes must be implemented on both Android and iOS unless explicitly scoped to one platform.
7. **Deployment Rule**: "Deploy" strictly means releasing directly to **Production** (`43.228.85.200`, Google Play Production track, App Store Production).
8. **Security**: Never commit `.env`, keystores, credentials, or binaries.

---

## 3. Status of Recent User Requests

| # | Request | Platform | Status | Resolution Details |
|---|---|---|---|---|
| **1** | Admin system to add new names & auto-generate meanings (`https://xn--b3cu8e7ah6h.com/admin`) | Backend/Web | ✅ Completed | Built `/admin/names` with automatic numerology sum (`เลขศาสตร์`), shadow sum (`พลังเงา`), 8-day Kalakini checks, and auto-meaning generation pipeline. |
| **2-3** | "เจเน่" typed but not ranking at #1 | Backend | ✅ Completed | Fixed search sorting/ranking to prioritize exact name match when user inputs query. |
| **4-5** | "ชื่อบุคคล" generic meaning issue (untranslatable vs error) | Backend | ✅ Completed | Implemented Option 1: normalized meanings to prioritize authentic etymology/Pali-Sanskrit roots over fallback placeholders. |
| **6** | When user taps a suggested name, auto-scroll to top | iOS / Android | ✅ Completed | Integrated `scrollToTopTrigger` in SwiftUI/Compose so selecting candidate names scrolls smoothly to wizard top. |
| **7** | Birthday selector screen jitter / shaking | iOS / Android | ✅ Completed | Stabilized dropdown selection state and layout sizing to eliminate viewport jumping on date change. |
| **11** | เมื่อเลือกวันเกิด แสดง UI การคัดกรองชื่อมงคลปลอดกาลกิณี | iOS / Android | ✅ Completed | Implemented compact single-row `KalakiniSummaryCard` in Step 2: shows shield icon with `"คัดเฉพาะชื่อมงคลปลอดกาลกิณี 100%"` (reduced sentence length by ~50% from original description), clean, punchy, and without prototype name or confusing alphabets. Verified on iOS simulator & Android build. |

---

## 4. Deep-Dive: iOS Mock Purchase Bug & Fixes (2026-09-14)

### The Issue
When user enabled "Mock purchase" via the bug icon on iOS, the suggested names container ("ตรวจคู่เลขชื่อที่ใกล้เคียง") was rendered as an empty green box instead of showing the 8 unlocked names with colored `PairCircles`.

### Root Causes & Fixes Applied:
1. **`KeychainHelper.swift` Save/Delete Overwrite Bug**:
   - *Bug*: In `save(key:value:)`, `UserDefaults.standard.set` was invoked *before* calling `delete(key:)`. However, `delete(key:)` removes the `UserDefaults` fallback key! This immediately wiped the newly stored entitlement token in simulator environments.
   - *Fix*: Moved `UserDefaults.standard.set(value, forKey: "keychain_fallback_" + key)` to run *after* `delete(key:)` and `SecItemAdd`.
2. **`SearchView.swift` Premature `onChange` Trigger**:
   - *Bug*: `.onChange(of: purchaseManager.effectiveIsUnlocked)` triggered immediately when `mockPurchaseEnabled = true` was set, *before* the asynchronous `verifyPurchase` network call completed and saved the token. The subsequent suggestion request was sent without `X-Entitlement-Token`.
   - *Fix*: Changed `.onChange` to observe `purchaseManager.isUnlocked`, which only turns `true` *after* `verifyPurchase` succeeds and the token is persisted.
3. **`SearchViewModel.swift` Task Cancellation Race**:
   - *Bug*: In `SearchViewModel.init`, the startup Task called `nameChanged()` shortly after `setMockPurchase`, triggering a debounced (280ms) `requestNameSuggestions(immediate: false)` that cancelled the in-flight priority `requestNameSuggestions(immediate: true)` request.
   - *Fix*: Introduced `suggestionIsImmediate` guard: if an immediate request is currently in-flight, debounced requests are skipped to prevent cancellation.
4. **Verification**:
   - Rebuilt with Xcode and launched on iPhone 17 Simulator (`CEABEFD0-B0C9-4003-8C81-D15CF9965FA9`) with arguments `-debugQuery จิรายุ -debugMockPurchase`.
   - Result: Response returned `data=8, unlocked=true`.
   - UI confirmed: Shows 8 unlocked names (`จิรายุ`, `จิรายุส์`, `จิรายุ์`, `จิรายุด`...) with colored `PairCircles` (`24/50`, `40/71`, `33/50`, `25/60`).
   - Cleaned up all temporary debug file-logging before handoff.

---

## 5. Key File Locations

### Go Backend
- `go-backend/delivery/http/handler/name_admin_handler.go`: Admin name management & creation.
- `go-backend/delivery/http/handler/templates/name_admin.html`: Name admin HTML template.
- `go-backend/usecase/name_admin_usecase.go`: Business logic for name insertion & meaning calculation.
- `go-backend/usecase/entitlement_service.go`: Google Play & App Store entitlement verification (`debug-mock-*` support).
- `go-backend/delivery/http/handler/name_handler.go`: `GET /api/v1/names/suggestions`.

### iOS Frontend (`ios-frontend/ChueDee/`)
- `Core/Networking/KeychainHelper.swift`: Keychain & UserDefaults fallback storage for tokens.
- `Core/Networking/APIClient.swift`: API networking client (`nameSuggestions`, `verifyPurchase`).
- `Core/Purchases/PurchaseManager.swift`: StoreKit 2 management & `DebugPurchaseView`.
- `Features/Search/SearchView.swift`: Main search wizard, step indicators, suggestions UI.
- `Features/Search/SearchViewModel.swift`: Search wizard state, suggestion tasks, and debouncing.

### Android Frontend (`android-frontend/app/src/main/java/com/miracle/names/`)
- `ui/wizard/SearchWizardScreen.kt`: Jetpack Compose wizard screen.
- `viewmodel/NameSearchViewModel.kt`: Search ViewModel.
- `billing/BillingManager.kt`: Google Play Billing and mock purchase handler.

---

## 6. Development & Verification Commands

### Go Backend
```bash
cd go-backend
go test ./...
go run main.go
```

### iOS Frontend (Simulator Build & Test)
```bash
cd ios-frontend
# Build for Simulator
xcodebuild -project ChueDee.xcodeproj -scheme ChueDee -sdk iphonesimulator \
  -destination "id=CEABEFD0-B0C9-4003-8C81-D15CF9965FA9" \
  -configuration Debug CODE_SIGNING_ALLOWED=NO build

# Install & Launch
xcrun simctl install CEABEFD0-B0C9-4003-8C81-D15CF9965FA9 \
  /Users/tayap/Library/Developer/Xcode/DerivedData/ChueDee-hkkittxymghabmawkwveshcmaovq/Build/Products/Debug-iphonesimulator/ChueDee.app

xcrun simctl launch CEABEFD0-B0C9-4003-8C81-D15CF9965FA9 com.numberniceic.app \
  -debugQuery จิรายุ -debugMockPurchase

# Screenshot capture
xcrun simctl io CEABEFD0-B0C9-4003-8C81-D15CF9965FA9 screenshot /tmp/sim_screen.png
```

### Android Frontend
```bash
cd android-frontend
./gradlew :app:assembleDebug
./gradlew test
```

### Server Deployment Information
- **Production Server IP**: `43.228.85.200`
- **Domain**: `https://ชื่อดี.com` / Punycode `https://xn--b3cu8e7ah6h.com`
- **Remote Service**: `systemctl status pj-names-miracle.service`
- **Remote Directory**: `/opt/pj-names-miracle`
- **Binary**: `/opt/pj-names-miracle/pj-names-miracle-server`
- **Environment File**: `/etc/pj-names-miracle/pj-names-miracle.env`
- **Local Upstream**: `127.0.0.1:18080`
- **Rule**: Always create a timestamped backup before updating production binaries.

---

## 7. Next Steps for Incoming AI Model / Developer
1. If the user asks to continue or deploy backend updates: compile for Linux (`GOOS=linux GOARCH=amd64 go build -o pj-names-backend-linux main.go`) and deploy via SSH to `43.228.85.200`.
2. If working on iOS/Android releases: run test suites first, then follow direct-to-production submission rules.
3. All mock purchase and suggestions unlocking features are fully working and verified.

---

## 8. Session Summary: 2026-09-26 Production Release & Polish

### 8.1 UI Refinements & Parity
1. **Centered Name Meaning Typography**:
   - `android-frontend/app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt`: Centered `displayMeaning` and `displaySurnameMeaning` along X axis (`textAlign = TextAlign.Center`, `fillMaxWidth()`).
   - `ios-frontend/ChueDee/Features/Search/SearchView.swift`: Centered `displayMeaning` in quotes with `.multilineTextAlignment(.center)` and `.frame(maxWidth: .infinity)`.
2. **iOS Auspicious Calendar Year Display (พ.ศ.)**:
   - `ios-frontend/ChueDee/Features/Auspicious/AuspiciousCalendarView.swift`: In `ThaiDateField`, adjusted spacing to 8, icon to 34x34, font to `sarabun(14, bold)`, horizontal padding to 10, and added `minimumScaleFactor(0.60)` to guarantee the full 4-digit Buddhist Era year (e.g. `2569`) renders completely without truncation.
3. **iOS 4D Diamond Profile Radar Edge-to-Edge Expansion**:
   - `ios-frontend/ChueDee/Features/Search/SearchView.swift`: Reduced outer padding in `decodedAnalysisSection` to `0pt` and inner padding to `6pt`, removed loose Spacers in `NameProfile4DCard` so the 4D radar chart spans edge-to-edge.
4. **Android Step Bar Glowing Halo Rendering**:
   - `android-frontend/app/src/main/java/com/miracle/names/ui/wizard/SearchWizardScreen.kt`: In `WizardStepBar`, reorganized `drawWithContent` so connector lines and the animated comet are drawn *before* step circles (`drawContent()`), ensuring the active step glowing aura (42dp) and corona ring (28dp) sit properly on top of connecting lines.
5. **Animation Restraint**:
   - Removed unnecessary drifting/moving UI elements across wizard screens, preserving only the intentional shining glow aura on numbers for a refined, professional look.

### 8.2 Production Deployments & Store Submissions
1. **Go Backend & Landing Page Live Deployment (`43.228.85.200`)**:
   - Replaced showcase screenshot in `go-backend/images/app-screenshots/ios-search-showcase.png` with new UI screenshot (`532x1024`).
   - Updated `go-backend/static/index.html` to reference `ios-search-showcase.png?v=3` for CDN cache-busting.
   - Cross-compiled binary, created server backup (`pj-names-miracle-server.bak-20260926112331`), synced assets, restarted `pj-names-miracle.service`, and verified live HTTP 200 OK.
2. **iOS App Store Release (`1.0.11`, Build `21`)**:
   - Bumped `MARKETING_VERSION = 1.0.11` and `CURRENT_PROJECT_VERSION = 21` in `ios-frontend/ChueDee.xcodeproj/project.pbxproj`.
   - Verified Debug build: `xcodebuild -project ChueDee.xcodeproj -scheme ChueDee -sdk iphoneos -configuration Debug CODE_SIGNING_ALLOWED=NO build` -> **BUILD SUCCEEDED**.
   - Built Release Archive: `build/ChueDee.xcarchive` -> **ARCHIVE SUCCEEDED**.
   - Exported and uploaded directly to **App Store Connect** -> **Upload succeeded** -> **Waiting for Review**.
3. **Android Google Play Release (`1.4.14`, versionCode `31`)**:
   - Bumped `versionCode = 31` and `versionName = "1.4.14"` in `android-frontend/app/build.gradle.kts`.
   - Built and signed release bundle via `./gradlew :app:bundleRelease` using upload keystore -> **BUILD SUCCESSFUL**.
   - Production bundle archived at `android-frontend/play-store-release/ChueDee-1.4.14-build31.aab` (30 MB).
   - Uploaded to Google Play Console Production track -> **Changes in review**.
4. **App Store Marketing Showcase Graphics**:
   - Generated 3 high-resolution App Store screenshots (`1242 x 2688 px`) with emerald-gold luxury theme, frosted glass cards, Prompt font typography, and realistic device mockups:
     - `iphone-65-01.png` / `appstore-marketing-01.png`: "ชื่อดี 4 มิติ"
     - `iphone-65-02.png` / `appstore-marketing-02.png`: "ปรับสูตรมงคล 4 มิติ"
     - `iphone-65-03.png` / `appstore-marketing-03.png`: "การันตีมงคล 100%"
   - Script saved at `ios-frontend/AppStoreScreenshots/marketing/generate_appstore_screenshots.py`.
5. **App Logo Update**:
   - Android launcher icon updated to new luxury gold on dark background inside bundle (`mipmap-*/ic_launcher.png`).
   - Store listing icon `app_logo-512.png` (512x512 px) uploaded to Google Play Console Main store listing.
6. **Web Article Typography Update**:
   - `go-backend/delivery/http/handler/templates/article_detail.html`: Imported Google Font `Sarabun` (weights 300..700). Applied `font-family: 'Sarabun', system-ui, -apple-system, sans-serif;` to all non-heading elements (paragraphs, excerpts, author bio, dates, breadcrumbs, editorial notes). Headings (`h1`–`h6`) strictly retain `Prompt` / `Sukhumvit Set`.
   - Updated `article-download-cta.css` related card tags to `Sarabun`.
   - Deployed and verified live on production server `43.228.85.200`.

