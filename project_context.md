# Project Context: Auspicious Name Analysis & Signature Design

This document serves as the Single Source of Truth (SSOT) for the project's architecture, tech stack, development guidelines, and current status.

## 1. Project Scope
A full-stack application focused on two main features:
- **Auspicious Name Analysis**: Semantic search and analysis utilizing a database of ~300,000 entries.
- **Custom Signature Design**: A feature to generate, design, and manage custom signatures.

## 2. Tech Stack & Core Technologies

### Frontend
- **Platforms**: Native Android in production; native iOS SwiftUI client in development
- **App Name / Store Title**: `ชื่อดี 4 มิติ` (Internal `app_name`: `ชื่อดี 4 มิติ`)
- **Application ID / Package Name**: `com.numberniceic.app` (Store Console: "ชื่อดี 4 มิติ")
- **Current Version**:
  - Android: `versionCode = 28`, `versionName = "1.4.12"` (Google Play release bundle targets Android 15 / API 36)
  - iOS App Store update prepared locally: `MARKETING_VERSION = 1.0.10`, `CURRENT_PROJECT_VERSION = 20` (iOS 16+); signed IPA is built but not uploaded.
- **Play Billing Release Baseline**: Production builds use Google Play Billing Library `8.3.0`. Debug builds follow the same server-verification rule; billing failures, missing product details, and restore requests never grant premium entitlement by themselves.
- **In-App Purchase Product ID**: `premium_lifetime` (lifetime access; display price always comes from the active store listing)
- **Language**: Kotlin
- **UI Framework**: Jetpack Compose (Exclusively)
  - *Rule*: XML layouts are strictly prohibited unless absolutely necessary for specific legacy Native APIs.

### Backend
- **Language**: Go (Golang)
- **Framework/Structure**: Clean Architecture (evident from `cmd`, `config`, `delivery`, `domain`, `repository`, `usecase` directories).
- **Role**: Handles all core logic, heavy architectural calculations, and data processing.

### Database
- **Primary Database**: PostgreSQL
- **Search Capabilities**: Configured for high-performance search, incorporating Vector Search capabilities (e.g., `pgvector` or `pg_trgm`) for meaning-based and semantic queries over the large dataset.

## 3. System Architecture & Data Flow

**Hub & Spoke / Go API to Kotlin UI Model**
- **Heavy Backend / Lightweight Frontend**: The Go backend acts as the central hub, performing all complex business logic, semantic search querying, and data processing.
- **API Contracts**: The frontend communicates with the backend via robust API contracts (REST).
- **UI State Management**: The Kotlin/Compose frontend remains purely lightweight, focusing solely on UI state management (MVI/MVVM), user interactions, and presenting the processed data received from the backend.

## 4. Directory & Folder Structure

- `/android-frontend/`: Contains the native Android application built with Gradle.
- `/ios-frontend/`: Contains the native iOS 16+ SwiftUI/MVVM application. It renders API state only; ranking and numerology business rules remain in Go.
- `/go-backend/`: Contains the Go application structured following Clean Architecture.

## 5. Development Strategy & Rules

- **UX & Search Mental Model**:
  - Step 1 strictly controls **Calculation Scope** with 2 options: **เฉพาะชื่อตัว (Standalone Name)** vs **รวมชื่อ + นามสกุล (Name + Surname)**.
  - Step 2 Day of Birth selection utilizes an `ExposedDropdownMenuBox` (Dropdown List) to optimize screen space.
  - If the prototype name has a decoded meaning, preselect that meaning as the default wish query on both Android and iOS; clear stale selections when switching prototypes.
  - Global Ranking Weight Ratio (`⭐ เลขศาสตร์ X% 🎯 ความหมาย Y% 🎶 เสียงชื่อ Z%`) is displayed once in the compact pinned results header, not repeated on every ranking row.
  - Ranking results should use a compact full-width table layout, not large cards. Each row has no rounded corners, uses thin row separators, and is divided into columns for rank, name, auspicious number pairs, score, and navigation affordance. Number pairs must be their own column. Include a sticky column header row (`#`, `ชื่อ`, `คู่เลข`, `คะแนน`) so users understand the table while scrolling. The rightmost action affordance should be an icon-only chevron in its own trailing space; avoid repeated action text such as `ดู` in every row because it reads like table data. The whole row remains clickable. Do not show meaning text in the row; users are scanning names first and can open details when needed.
  - Ranking uses one space-efficient control strip above the sticky header: formula posture plus in-memory filters (`ทั้งหมด`, `คู่ดี`, `ควรพิจารณา`). Caution rows alone receive one compact explanation line naming their strongest non-numerology dimension and R5/R7 risk. Reranks briefly show `ใหม่` or an up/down rank movement; tapping any pair circle opens its dictionary detail. These interactions must not trigger another ranking API request.
  - `names_miracle.thname` is the governed shared word corpus for both first-name and surname candidates; raw storage does not split them into separate columns. In combined mode a compact toggle selects `search_target=first_name` (`[หาชื่อตัว] + นามสกุล`) or `search_target=surname` (`ชื่อตัว + [หานามสกุล]`). The user-entered part is fixed and is never rejected or altered, even when it contains Kalakini characters. Kalakini filtering applies only to the searched candidate. The fixed part is also sent as `prototype_name`, so `pg_trgm` deliberately surfaces counterpart candidates with similar spelling or pronunciation for the user to consider. Thai full-name display always keeps the first name before the surname: when finding a surname, the fixed user-entered first name is rendered first and the candidate surname second. Every row and summary must label the candidate role. Locked rows mask only the searched candidate portion. Standalone mode never appends hidden combined state.
  - Switching the combined search target clears the fixed-part input because its semantic role changes between surname and first name; stale text must never be silently reinterpreted. Search captures the selected target before the network request and stores that captured value in the result snapshot so late UI changes cannot relabel returned rows.
  - Unnecessary "2-syllable filter" UI checkbox removed.
  - All names (1-syllable or 2+ syllables) compete naturally in the ranking based purely on their astrological and semantic scores without artificial rejection.
- **Compose Standards**: All new UI components must be built using Jetpack Compose following modern state hoisting and unidirectional data flow principles. Reusable composables should be prioritized.
- **Typography**:
  - **Android**: Uses Sarabun as the primary app font through Compose `MaterialTheme`. Number-pair circles (`PairCircles`) use `FontFamily.Default` for digit compactness.
  - **Web**: Uses `'Sukhumvit Set'` prioritized in the font stack (`'Sukhumvit Set', 'Prompt', 'Sarabun', system-ui, -apple-system, sans-serif`) across all article titles, headings, and article cards for Apple typography on macOS/iOS, with graceful Prompt/Sarabun fallbacks for other platforms.
- **Thai Buddha Assets**: Blessing cards use Photoshop-style transparent PNG assets from the backend static route `/buddhas/{day}_buddha.png`, then fall back to bundled Android `drawable-nodpi` PNG resources when the dev backend is offline. Do not fall back to SVG for this screen; PNG is the visual direction because it is easier to compose as a top-right wallpaper ornament. During development these Coil requests disable memory and disk cache, while keeping network fetching enabled and adding a timestamp query string so replacing image files in `go-backend/thai_buddha/` is reflected after the Android screen reloads. Do not disable Coil network cache policy here, because that can prevent remote images from loading. When adding the remaining weekday Buddha images, add backend PNG files first, then matching Android fallback PNG resources.
- **Wednesday Buddha Naming**: Use `wednesday1_buddha.png` for Wednesday daytime and `wednesday2_buddha.png` for Wednesday night. Android day keys are `wednesday1` and `wednesday2`; keep backend PNG filenames and Android `drawable-nodpi` fallback resource names aligned with those keys.
- **Blessing Card Visual Direction**: The blessing card is intended to be captured and reused as a phone wallpaper. Use a gold-card visual direction for the captured card itself: warm gold gradient background, dark purple/brown text for legibility, Buddha artwork placed at the top-right as a decorative watermark/ornament, centered narrow text column, and numerology meanings below their pair circles.
- **Terminology Standards**:
  - Main Vowel/Auspicious Number Sum: **เลขศาสตร์**
  - Consonant/Shadow Number Sum: **พลังเงา**
- **UI Component Standards**:
  - All number pairs (คู่เลข) presented across the app (both in search results and Blessing Cards) MUST be rendered as colored circles (`PairCircles`) indicating their auspiciousness level, rather than plain text brackets.
  - Pair detail bottom sheets must also use the shared `PairCircles` component. Do not draw one-off number badges in detail sheets; visual consistency across ranking rows, blessing cards, decode panels, and pair detail sheets is a product rule.
  - The active brand direction is **Louis Vuitton Luxury Palette (เอสเปรสโซ่เข้ม / น้ำตาลอบอุ่น / ดาร์กช็อกโกแลต + ทองคำบริสุทธิ์)**: `#140B07` (Deep Espresso Base), `#26160E` (Warm Coffee Header / Accent), `#160C08` (Dark Chocolate Cards), `#D4AF37` / `#F3E5AB` (Warm Champagne Gold). Applied consistently across Android (Compose), iOS (SwiftUI), and Web Backend (CSS / HTML / Admin / Customer Chat / Articles).
  - Important purchase/unlock copy currently uses: `100 อันดับชื่อที่ดีที่สุด`, `ชื่อดีไม่ได้เกิดจากความบังเอิญ`, button `เปิดชื่อทั้งหมด`, and a price badge populated from Google Play/App Store. Keep the tone confident and non-pushy; never hard-code a currency amount in native UI.
- **Backend Development & Deployment Workflow**: All backend development must be performed and verified locally first (`go test ./...`), then cross-compiled and deployed to the production server (`43.228.85.200`).
- **Deployment Rule**: "Deploy" strictly means releasing/submitting directly to **Production** (live for real end-users on Google Play Production track & App Store Production), never internal/testing tracks unless explicitly requested by the user.
- **Native Client API Contract**: Android and iOS use `/api/v1` and must verify `X-API-Version: 1` plus `X-Ranking-Contract-Version: adaptive_pairs_v1`. The same ranking contract version is returned inside `policy`. `GET /api/v1/readiness` verifies the database and warmed sound profile and exposes approval status; clients retry a 503 rather than treating it as an empty ranking. The canonical schema is `go-backend/docs/openapi-v1.yaml`.
- **Web Portal**: Includes `/`, `/articles`, `/about` (ความเป็นมาและแรงบันดาลใจ), `/methodology`, `/isan-auspicious-method` (เอกสารหลักการดูฤกษ์ยามตามตำราอีสานฉบับอ่านง่าย), `/privacy`, `/terms`, `/contact`, `/download`, `/chat`, and `/auspicious-calendar`.
- **Auspicious Date Finder**: `/auspicious-calendar` begins with the user's activity and practical date range instead of a raw calendar. Database-backed category tabs keep the growing activity list practical: ชื่อและเอกสาร, งานและธุรกิจ, บ้านและทรัพย์สิน, ครอบครัวและพิธี, and การเดินทาง. The catalog includes brand/business naming, important-document filing, product/website launches, and career starts in addition to the original activities. `GET /api/v1/auspicious-activities` supplies the choices and `POST /api/v1/auspicious-search` evaluates up to 366 days in Go, returning at most three explainable recommendations. Every matched rule returns its database-backed short meaning for an expandable explanation in the web result. Gregorian-to-Thai-lunar conversion and วันพระ come from `thailunar`; versioned rules and internal source provenance live in PostgreSQL migration `000038_auspicious_rules`. Conflict policy is deterministic: any exclusion wins, and overlapping positive labels in the same signal group contribute only the highest-priority score. The user-facing rule-set label is `ฤกษ์วันมงคลตามตำราอีสาน`; a specific website name is not shown to users, while provenance is retained internally for rule review.
  - **Native Taksa analysis board**: Android and iOS present the Taksa wheel as a read-only analysis board after birth date/time entry. Users may tap any of its eight sectors to open a bottom sheet showing the backend-provided role meaning and how that weekday affects their auspicious-date analysis; the center cell is not selectable. Role interpretation and personalized auspicious guidance remain Go-owned API data.
  - **Admin-managed auspicious content (2026-08-28)**: The existing Basic-Auth admin dashboard links to `/admin/auspicious-content`. Administrators can edit user-facing rule-set names, activity names/icons/descriptions/scope and safety notes, auspicious-rule titles/descriptions/meanings, all eight Taksa role names/meanings/wheel descriptions, dynamic Taksa sentence templates, and the search disclaimer. Migration `000053_auspicious_admin_content` stores the new text catalog in PostgreSQL. Go reads current content for each search/Taksa request so web, Android, and iOS receive updates immediately without a client rebuild; scores, rule conditions, progression, and other business calculations are deliberately not editable as content.
  - **2026-08-28 Admin-managed Auspicious Content Production Release**: Applied migration `000053`, deployed backend SHA-256 `876a4dcfb3d68113a20ef3169b224034448ef92520cd18ecc48f12afd0b5c6fa`, and verified the live HTTPS health/API contract, Taksa calculation, database-backed meanings, authenticated admin page, and dashboard link. The pre-release binary and auspicious-table dump are stored at `/opt/pj-names-miracle/backups/deploy-20260828-084706/`. Android Debug/Release and iOS Debug/Release API base URLs are the Production HTTPS host; the native clients do not require a rebuild for admin content changes.
  - Every auspicious-date result must also identify major Buddhist holy days when applicable (มาฆบูชา, วิสาขบูชา, อาสาฬหบูชา, เข้าพรรษา, ออกพรรษา) alongside วันพระ/lunar metadata, so users can make an informed choice. This label is informational and does not exclude or alter the score unless the user enables `หลีกเลี่ยงวันพระ` or an activity rule excludes it.
- **Auspicious Activity Catalog**: `เปลี่ยนชื่อ ตั้งชื่อสกุล หรือยื่นเอกสารสำคัญ` is the primary activity because the web portal promotes the naming app; it absorbs the previously separate important-document option. `เปิดกิจการหรือเปิดตัวสินค้า` likewise covers store/business opening plus product, website, app, page, and online-store launches. The redundant `important_documents` and `product_launch` rows remain stored but disabled by migration `000044`. The catalog also includes receiving a new vehicle, childbirth, wedding, housewarming, ordination, brand naming, and starting a new job. UI text must distinguish user convenience preferences from source-authored rules and must disclose when an activity currently uses only general rules. Childbirth results always state that medical safety and clinician advice override auspicious timing.
  - **Negotiation & contracts (2026-08-30)**: `เจรจาต่อรอง ทำสัญญา หรือปิดการขาย` is a work-and-business activity. It combines the existing Isan general rules and commerce-oriented ปุณณดิถี with the day-level Thai กาลโยค rule set. กาลโยค uses Chula Sakarat, changes on 16 April in the simplified calendar convention, boosts วันอธิบดี and วันธงชัย, and exposes วันอุบาทว์/วันโลกาวินาศ as caution signals rather than hard exclusions so users can see overlapping traditional signals. It is not a time-of-day electional chart; a future ฤกษ์บน feature must return explicit time windows. Migration `000055` and backend SHA-256 `9e2f324bda2c65e26c216813b3fc1da681706153f47e95a9a6fa2500a6f9a702` were deployed to Production and verified over the public HTTPS API.
  - **Outfit colour advice (2026-08-30)**: Each personalized auspicious-date recommendation includes display-ready Thai outfit colours from the legacy Nab Ha Sri logic. The Go API calculates supporting and caution sets for the appointment day, the user's astrological birth day, and progressed age. It preserves the old 9-slot palette, evaluates กาลีก่อน, and falls back from ศรี to มนตรี whenever an auspicious set would collide with a caution or prior supporting set. กาลี has priority: if มนตรี still collides, that supporting set is omitted rather than inventing another colour, so users never receive contradictory advice. API sets carry their source context, Taksa role and meaning, planet name, number, and three display colours; those fields remain available for a future detail view.
    - **Compact mobile presentation**: The default clothing-colour card must show exactly two uncluttered rows only: `สีที่เสริม` and `สีที่ควรเลี่ยง`, with the backend-returned colour circles grouped in source order. Do not display a card title, source labels, planet names, role names, or explanatory copy in this compact card. Android and iOS only render the API response and never recalculate or replace colours locally.
- **2026-09-14 4D Name Search Latency & PostgreSQL Trigram GIN Optimization**:
  - **Issue**: Searching on Step 4 with prototype name (`prototype_name`) and wish query (`wish_query`) hung on `MagicalLoadingView` for **65.687 seconds** on first request (cold cache) and ~10 seconds on subsequent requests.
  - **Root Cause**: `SearchSemanticAndSimilarNames` in `go-backend/repository/name_repository.go` omitted the trigram similarity filter from its `WHERE` clause. When constructing the full rerank candidate pool with `limit = 0`, PostgreSQL executed a sequential/heap scan over all 92,207 candidates (or up to 250,000+ candidates depending on day of birth) and de-TOASTed over 6 GB of 1536-dimensional vectors from disk, causing massive disk I/O thrashing on VPS.
  - **Fix & Index Optimization**:
    - Added `(thname % $4 OR normalize_thai_phonetic(thname) % $5)` into the `WHERE` clause of `SearchSemanticAndSimilarNames` and `SearchSimilarNames`.
    - Leveraged GIN trigram indexes `idx_names_miracle_thname_trgm` and `idx_names_miracle_phonetic_trgm` via `BitmapAnd` with `idx_names_miracle_sha_sum_len`.
    - Configured `ALTER DATABASE tayap SET pg_trgm.similarity_threshold = 0.15;` on production PostgreSQL and executed `SELECT set_limit(0.15)` on connection pool initialization in `config/database.go` to match the 0.15 similarity threshold contract.
  - **Performance Benchmark Results**:
    - Candidate retrieval SQL duration dropped from **65,687ms** to **21.8ms** (~3,000x faster).
    - End-to-end first request latency (cold cache + OpenAI embedding generation) dropped from **66.0s** to **~1.1s** (~60x faster).
    - End-to-end repeated/cached vector search dropped from **10.3s** to **~280ms** (~35x faster).
  - **Step 4 Dynamic 4D Radar Graph Weight Percentages (2026-09-14)**:
    - Fixed an issue where the left-side percentage indicators (`⭐ เลขศาสตร์`, `🎯 ความหมาย`, `🔍 เอกลักษณ์`, `🎶 ไพเราะ`) on Step 4 ("ยืนยันโปรไฟล์ 4 มิติ & ค้นหาชื่อมงคล") did not change when dragging the point on the diamond radar selector (`RadarWeightSelector`).
    - **Root Cause**: The UI was previously reading from static `decodedProfile4D` instead of active ranking weights (`viewModel.weights` on iOS, `pendingRankingWeights` on Android).
    - **Fix**: Updated both iOS (`SearchView.swift`) and Android (`SearchWizardScreen.kt`) to compute the 4 percentages dynamically in real time from the active ranking weights using the largest remainder method (Hare-Niemeyer), ensuring they always sum exactly to 100%.
  - **Step 2 Birthday Selection UX**: Refined Step 2 birthday selection UI to focus on clarity: displays a concise single-line description explaining that day-of-birth selection filters premium non-Kalakini names for the next step, without showing confusing character/vowel lists or prototype name references.
  - **Step 3 Meaning Auto-Play & Media Controls UX/UI (2026-09-14)**:
    - Redesigned Step 3 ("เสริมความหมายมงคลที่ต้องการ") across Go backend, Android, and iOS.
    - **Go Backend API**: Raised `limit` threshold to 108 in `delivery/http/handler/name_handler.go` and `repository/name_repository.go` (`GET /api/v1/meanings/random?limit=108`). Cross-compiled and deployed live to production server `43.228.85.200`.
    - **Exact 36 Sets (108 Meanings Pool)**: Frontend fetches 108 random auspicious meanings, strictly configured as exactly 36 sets of 3 items (`ชุดที่ 1/36` to `ชุดที่ 36/36`).
    - **Removed Top-Right Refresh Button**: Streamlined header UI by removing the redundant top-right refresh icon, giving full attention to the auto-play controls and pinned meaning.
    - **8-Second Auto-Play**: Automatically cycles to the next 3 meanings every 8 seconds with smooth fade and gold flash animations (`LaunchedEffect(isMeaningAutoPlaying)` on Android / `Timer.publish(every: 8)` on iOS).
    - **Media Control Strip**: Compact bottom control bar featuring Backward (`⏮`), Play/Pause (`▶` / `⏸` with toggle label "พัก (8s)" / "เล่นต่อ"), Forward (`⏭`), and page counter (`ชุดที่ X/36`).
    - **Interactive Pinning**: Tapping any meaning card immediately pins it to the active intent header box.
    - **Fixed 2-Line Row Height & Compact Vertical Spacing**: Both header box and all 3 suggestion rows enforce fixed 2-line vertical reservation (`minHeight = 44.dp` / `46.dp`, vertical padding reduced to 6–10dp) so that cycling between 1-line and 2-line meanings keeps the container height completely stable without causing the screen to jump or shift.
- **2026-09-26 Mobile Production Release (iOS 1.0.11 / Android 1.4.14)**:
  - **iOS (`com.numberniceic.app` v1.0.11, Build 21)**: Built archive `ChueDee.xcarchive` and successfully exported/uploaded to App Store Connect. Includes Buddhist Era year display fix in calendar picker, edge-to-edge 4D profile radar layout, centered meaning text, and shining number aura effect.
  - **Android (`com.numberniceic.app` v1.4.14, versionCode 31)**: Built and signed production bundle `ChueDee-1.4.14-build31.aab` (saved in `android-frontend/play-store-release/`). Includes glowing halo aura rendered above connector lines and centered meaning text.
  - **Web Landing Page & Backend**: Deployed to production server `43.228.85.200` with updated showcase screenshot (`v=3`).
- **2026-09-27 Support Chat Admin Message Deletion**:
  - Administrators can permanently delete an individual customer or อ.ทญา message from `/admin/chat`; the web page asks for confirmation before deletion.
  - The authenticated admin iOS app supports the same operation by long-pressing a message and choosing `ลบข้อความ`.
  - Deletion is performed on the shared backend record, recalculates the conversation preview and unread counters, and removes an associated upload when possible. Customer Android/iOS clients and the admin clients remove the message on their next message refresh.
  - Production backend binary SHA-256 `5a574cbf1a8d196bb957dcc9c47e079c51f62e91baeee36df3b98bff2251f169` was deployed to `43.228.85.200` after backing up the prior binary; `GET /api/v1/health` returned healthy.
- **2026-09-13 Name Suggestions & Entitlement v2 Production Release**:
  - **Go Backend (`43.228.85.200`)**: Deployed binary SHA-256 `a924b8fd4306c9785d7d0dd13f4956f338ab6ca81192a142c9bad3f4a7256557`. Implemented `GET /api/v1/names/suggestions` with PostgreSQL `pg_trgm` similarity ranking constrained to green numerology (`เลขศาสตร์`) and shadow energy (`พลังเงา`). Free users receive name count and up to 2 preview number pairs without exposing real names or IDs (`premium_unlocked: false`). Paying customers receive full unmasked suggestions (`premium_unlocked: true`).
  - **Store Entitlement Verification**: Upgraded entitlement token to Version 2 with 10-year lifetime validity. Maintained resilient store fallback for native client-verified purchases (Android Google Play Billing & iOS App Store StoreKit 2) while enforcing strict fail-closed rejection on invalid or empty tokens. Backward-compatible with legacy Version 1 tokens for name search ranking.
- **2026-08-30 Wedding Riang Mon, Taksa Outfit Advice & Mobile Production Release**:
  - **Go Backend & Database**: Applied migration `000056_wedding_riang_mon` on Production database `43.228.85.200`. Added rule set `thai_wedding_tradition_v1`, rule `thai_wedding_riang_mon` (+25 points, Priority 500) covering Waxing 7, 10, 13 and Waning 4, 8, 10, 14, and updated wedding ranking priority in `auspicious_service.go` to always prefer Riang Mon dates. Linked Kala Yok signals to wedding.
  - **Android (`com.numberniceic.app` v1.4.6, versionCode 21)**: Built and uploaded signed bundle `chuedee-1.4.6-vc21.aab` to Google Play Console Production track. Combined birth date and birth time into a unified luxury card with auto-scroll to the magical purple time field, borderless Taksa wheel, Taksa outfit color guidance with downward bracket, and global floating support chat.
  - **iOS (`com.numberniceic.app` v1.0.7, Build 14)**: Built archive and submitted to App Store Connect Production ("Waiting for Review"). Parity with Android for Taksa wheel, wedding Riang Mon, outfit color advice, unified birth card, and global floating support chat.
- **2026-08-26 Auspicious Finder Production Release**: Deployed migrations `000038`–`000044`, the `/auspicious-calendar` finder, `/isan-auspicious-method` reference page, public footer links, and the simplified 9-activity catalog to Production. The live API search was verified against a 30-day range and returned three explained recommendations under `ฤกษ์วันมงคลตามตำราอีสาน`. The pre-release binary and public static pages are backed up at `/opt/pj-names-miracle/backups/deploy-20260826-071100/`.
- **2026-08-26 District Name-Change Article Production Release**: Applied migration `000045` to publish `เปลี่ยนชื่อที่อำเภอต้องเตรียมอะไรบ้าง?`. Verified the live article, article index card, app CTA, and official DOPA source link over HTTPS. The pre-migration `articles` data backup is `/opt/pj-names-miracle/backups/deploy-20260826-072724/articles-before-000045.sql`.
- **2026-08-26 Full SEO, Schema & GA4 Production Release**: Deployed complete SEO suite to Production server `43.228.85.200`. Deployed updated dynamic XML sitemap (including `/privacy`, `/terms`, and all dynamic articles), comprehensive 301 permanent redirects for `.html` aliases, site-wide Google Analytics 4 (GA4) across all public pages & article templates, and Schema.org JSON-LD structured data with breadcrumbs. Verified live endpoints and GA4 tracking tags. Pre-release backup at `/opt/pj-names-miracle/backups/deploy-20260826-081835/`.
- **2026-08-16 Support Chat & Web Updates**:
  - Go backend + PostgreSQL support chat deployed on production (`43.228.85.200`).
  - Customer web `/chat` and backend admin `/admin/chat` are fully operational with real-time polling and lightbox image modals.
  - 5 Auspicious 3D animal avatars implemented across all platforms (`cat`, `dog`, `rabbit`, `bear`, `panda`) in a clean single row without gender constraints or text labels.
  - **Dynamic Backend-Hosted Avatars**: All avatars are served dynamically from the backend server (`/opt/pj-names-miracle/images/`). Avatars are optimized (~80KB at 256x256) and cached synchronously in memory on iOS (`AvatarImageCache` / `NSCache`) and Android (Coil memory/disk cache). Changing avatar images on the backend takes effect immediately on clients without requiring frontend app rebuilds or redeployments.
  - Unread notification badges on Android (Floating Action Button with red badge and count) and iOS (Top Bar Chat Icon with red capsule badge and count).
  - Backend endpoint `GET /api/v1/chat/conversations/:id/unread` provides lightweight polling for unread counts, and opening chat automatically marks customer messages as read.
  - **Read Receipts ("อ่านแล้ว")**: Implemented across all platforms (Web Customer, Web Admin, Android, and iOS). Migration 34 added `is_read` column to `support_messages`.
  - Local verification commands:
    - Backend: `cd go-backend && go test ./...`
    - Android: `cd android-frontend && ./gradlew :app:assembleDebug`
   - iOS: `cd ios-frontend && xcodebuild -project ChueDee.xcodeproj -scheme ChueDee -destination 'generic/platform=iOS' -configuration Debug CODE_SIGN_IDENTITY="" CODE_SIGNING_REQUIRED=NO CODE_SIGNING_ALLOWED=NO build`
- **2026-09-13 Collapsible App Bar (Android & iOS)**:
  - Both platforms implement a collapsible header: the top row (logo, "ชื่อดี 4 มิติ", action buttons) hides on scroll down; the `StepProgressBar`/`WizardStepBar` stays pinned. Scrolling back up restores the full header.
  - **Android**: Custom `Surface(color = GreenHeader)` with `AnimatedVisibility` (`expandVertically`/`shrinkVertically` + `fadeIn`/`fadeOut`) and `NestedScrollConnection` (threshold ±15f). `.statusBarsPadding()` prevents edge-to-edge clipping.
  - **iOS**: `ScrollOffsetObserver` (`UIViewRepresentable`) uses UIKit KVO on `UIScrollView.contentOffset` for guaranteed per-frame scroll callbacks. Header is a `VStack` child above `ScrollViewReader`, **not** inside `.safeAreaInset(edge: .top)`. Uses `.background(Color.brandGreen.ignoresSafeArea(edges: .top))` to extend green behind status bar.
  - **Critical SwiftUI caveat**: SwiftUI `PreferenceKey` + `GeometryReader` inside `ScrollView` does NOT fire `onPreferenceChange` on every scroll frame on iOS 16; it batches updates and may fire only when scrolling stops. `.safeAreaInset(edge: .top)` cannot be used for collapsible headers because height changes cause a layout feedback loop. The reliable solution is UIKit KVO on the underlying `UIScrollView.contentOffset`.

- **Logic Separation**: No business logic or complex data transformations should exist in the Android frontend. If a calculation is needed, it must be migrated to the Go backend and exposed via API.

## 6. Current Project Status & Immediate Next Steps

- **Current Status**: Exhaustive, auditable ranking is implemented end to end. Backend golden tests are independently checked against read-only SQL oracles; Android freezes the applied search/rerank formula and provenance with each result set. Database governance SQL is staged locally but has not been applied to production.
- **2026-08-16 Support Chat Delivery & Handoff**:
  - **Production is live** on `https://ชื่อดี.com` / punycode host `xn--b3cu8e7ah6h.com`, deployed by SSH as `root@43.228.85.200`. Go service is `pj-names-miracle.service`, working directory `/opt/pj-names-miracle`, binary `/opt/pj-names-miracle/pj-names-miracle-server`, local upstream `127.0.0.1:18080`, and environment file `/etc/pj-names-miracle/pj-names-miracle.env`. Always make a timestamped binary backup before restarting the service.
  - **Database migrations applied locally and production**: `go-backend/db/migrations/000029_support_chat.up.sql` (conversations/messages) and `000030_support_chat_inbox.up.sql` (Inbox counters/preview). The tables are `support_conversations` and `support_messages`.
  - **Public customer web flow**: `/chat` is a real customer support page. It creates an anonymous room and stores `conversation_id` + opaque `customer_token` in browser localStorage (`chuedee_support_chat_v1`), so the same browser resumes its chat. It supports one JPG/PNG/WEBP image per message, max 8 MB.
  - **Admin flow**: `/admin/chat` uses the existing `ARTICLE_ADMIN_USERNAME` / `ARTICLE_ADMIN_PASSWORD` Basic Auth. Inbox tabs are `รอแอดมิน`, `กำลังคุย`, `ปิดแล้ว`; the server returns at most 50 rooms per view. A customer message reopens a closed room and increments `admin_unread_count`; an admin reply resets it. `/admin/chat/summary` supplies `{needs_reply, active, closed}`. The admin page polls every 5 seconds, updates tab badges/title, shows a toast, and can use Browser Notifications after the admin taps the permission button.
  - **Chat APIs**: `POST /api/v1/chat/conversations` with multipart `customer_name`; `GET|POST /api/v1/chat/conversations/{id}/messages` with `X-Chat-Token`. The post accepts multipart `body` and optional `image`. Android and iOS must persist the token in secure app storage; a token is the only customer identity in this MVP.
  - **Web discoverability**: homepage has a gold `ปรึกษาผู้เชี่ยวชาญ` navigation CTA, hero CTA, and floating CTA to `/chat`. Mobile nav moves that CTA to its own full row. The web bubbles use generated customer avatar `/images/chat_customer_avatar.png` and admin logo `/images/app_logo.png`; customer male/female assets are also deployed at `/images/chat_customer_female.png` and `/images/chat_customer_male.png`.
  - **2026-08-16 Cute Animal Avatars & Cross-Platform Support Chat (Gender-Neutral)**:
    - **5 Cute Animal Avatars**: Replaced binary gender selection with 5 auspicious, cute animal avatars: `cat` (แมว), `dog` (สุนัข), `rabbit` (กระต่าย), `bear` (หมี), and `panda` (แพนด้า) with smooth 3D claymation aesthetic on emerald & gold theme.
    - **Database Migration**: `go-backend/db/migrations/000032_support_chat_animal_avatars.up.sql` updated constraint `chk_support_conversations_avatar CHECK (customer_avatar IN ('cat', 'dog', 'rabbit', 'bear', 'panda', 'female', 'male'))` with default `'cat'`.
    - **Backend & Web**: Go backend normalizes, validates, and stores animal avatars; `/chat` public web client allows 5 animal selection and shows selected avatar on outgoing bubbles; `/admin/chat` renders the customer's chosen animal avatar across threads and active chats.
    - **Android Synchronization**: `SupportChatViewModel.kt`, `ChatAvatarHelper`, and `SupportChatScreen.kt` updated to display the 5 animal choices in a horizontal scrollable row and sync with the backend.
    - **iOS SwiftUI Native Client**: `SupportChatViewModel.swift`, `ChatAvatarOption`, and `SupportChatView.swift` updated with 5 animal choices in a horizontal ScrollView, Keychain persistence, full-screen chat UX, image attachment, and viewer. Verified with `xcodebuild` and deployed to production.
  - **Push Notification Architecture (FCM / APNs)**: Implemented end-to-end device token registration API (`POST /api/v1/chat/conversations/:id/device-token`), PostgreSQL migration `000034_support_chat_device_tokens.up.sql`, async Go dispatch service (`PushNotificationService`) triggered on admin replies, Android `SupportChatViewModel.registerDeviceToken`, and iOS `SupportChatViewModel.registerDeviceToken`.
  - **Not yet done**: Production FCM/APNs private keys injection in `.env`; true unread/read receipts; pagination/cursor beyond the current 50-row Inbox cap; secure encrypted Android token storage (current Android MVP uses ordinary `SharedPreferences`); grouping consecutive message bubbles to show one avatar per sender run; Android/iOS release deployment. Do not claim those are shipped.
- **2026-08-19 4D Dimension Labels, Settings Admin Backoffice & Morphological Identity Framework**:
  - **Single Name Analysis vs 4D Ranking Methodology**:
    - **Single Prototype Name Decode (การวิเคราะห์ชื่อเดี่ยว)**: Analyzes the intrinsic linguistic and astrological properties of a single name across 4 core dimensions:
      1. **ความหมาย (Meaning)**: Evaluates semantic depth and auspicious keywords from the ~300,000 dictionary corpus.
      2. **เลขศาสตร์ (Numerology)**: Calculates harmonic pair weights and shadow matrix scores (`เลขศาสตร์` & `พลังเงา`).
      3. **เอกลักษณ์ (Morphological & Root Identity)**: Measures the structural integrity of Pali-Sanskrit roots, character balance, and flexibility for auspicious expansion.
      4. **ไพเราะ (Sound / Phonetics)**: Measures syllable balance, vowel-consonant ratios, and euphony.
    - **4D Profile Decode Card (Hybrid UI with Normalized 100% Shares & Diamond Radar)**:
      - Displays on the left side the 4 dimensions normalized to 100% sum ($w_{\text{num}} + w_{\text{sem}} + w_{\text{res}} + w_{\text{snd}} = 100\%$) along with their quality label.
      - Displays on the right side an exact Diamond Radar Canvas (`NameProfileRadarView` on Android / `NameProfileRadarShapeView` on iOS) matching the ranking widget.
      - The center spark indicator dot position $(\text{SparkX}, \text{SparkY})$ must visually reflect the dominant dimension of the decoded prototype name on the 4-axis diamond plane ($X$: ความหมาย vs เอกลักษณ์, $Y$: เลขศาสตร์ vs ไพเราะ). Every diamond uses the same orientation: top = เลขศาสตร์, left = ความหมาย, right = เอกลักษณ์, bottom = ไพเราะ. Corner icons must sit outside the plot with a reserved visual gutter.
    - **4D Ranking Model (การจัดอันดับ 100 ชื่อที่ดีที่สุด)**: The 4th dimension acts as **ความคล้าย / Root Resonance (`pg_trgm`)** to identify candidates that preserve the phonetic and spelling silhouette of the user's prototype name.
    - **Combined Mode 4D Search Methodology (การค้นหาแบบผสมผสาน ชื่อ + นามสกุล)**:
      - **Baseline Profile Decode**: Decodes the input reference name (e.g. surname) first to establish the baseline 4D energy profile and pairs.
      - **Harmonic Search**: Dispatches candidate search for the target (e.g. first name) with weights initialized from the decoded reference profile.
      - **Full-Name Composite Calculation**: Evaluates the candidates against the surname across all dimensions: (1) Numerology sum (`ชื่อตัว + นามสกุล`) with no Kalakini on first name, (2) Joined semantic story, (3) Phonetic transition flow, and (4) Root compatibility.
  - **Admin Web Portal (`/admin/dimension-labels`)**: Live on `https://ชื่อดี.com/admin/dimension-labels` with emerald-gold theme, allowing admins to adjust minimum score thresholds (`min_score`), display labels (`label`), and detailed explanations for all 4 dimensions (`meaning`, `numerology`, `sound`, `resonance`).
  - **Database Migration**: `000036_dimension_labels.up.sql` applied to PostgreSQL database `tayap` on server `43.228.85.200`.
  - **Public Methodology Reference**: `https://ชื่อดี.com/methodology` updated with comprehensive explanations and interactive diamond radar chart for user reference.
  - **Dynamic In-Memory Caching**: `DimensionLabelRepository` in Go backend caches labels in memory with atomic invalidation on update, ensuring sub-millisecond response times for `/api/v1/decode`.
  - **Seamless Cross-Platform Sync**: Changing labels in the admin dashboard instantly affects the 4D Profile calculation for Android, iOS, and Web clients without requiring client updates or app store releases.
  - **Verification**: Verified with `go test ./...` in Go backend, `assembleDebug` in Android, and `xcodebuild` in iOS.
- **Recent Technical Implementations**:
  - **Dynamic Sample Names (Celebrities)**: `GetCelebrities` endpoint in Go backend now queries the `sample_names` database table dynamically (`SELECT name FROM sample_names`) via Clean Architecture repository/usecase layers instead of hardcoded lists.
  - **Full-Bleed Celebrity Marquee Layout**: Refactored `SearchWizardScreen` layout hierarchy by scoping horizontal padding to inner child cards rather than the parent scrollable column, allowing the celebrity `LazyRow` to span full screen width seamlessly without edge gaps.
  - **Single-Tap Gesture & Smooth Marquee**: Replaced continuous `animateScrollBy` with frame-based `dispatchRawDelta` via `withFrameNanos` to prevent `isScrollInProgress` from locking tap gestures. Avatar items now respond to a **single tap** to stop scrolling and trigger immediate name decoding.
  - **Wish Auto-Selection Logic**: Manual meaning input has been removed. If a decoded name has no database meaning, Android immediately requests DB-derived random meanings, pins the first as an explicit automatic recommendation, shows three alternatives, and lets the user refresh. The request is revision-guarded so a late response cannot overwrite a newly typed name. A built-in all-round wish remains the network-failure fallback. A real database meaning is never replaced by randomization.
  - **Search Mode UI Animations**: Implemented `AnimatedVisibility` (expandVertically + fadeIn) for mode toggles and analysis sections to ensure smooth, natural UX transitions when switching between Standalone and Combined name search modes.
  - **Database-Backed Wish Pinning & Auto-Refresh**: Wish text is loaded from distinct, non-empty `names_miracle.meaning` values through `GET /api/v1/meanings/random`; Android must not maintain a hard-coded wish pool. The decoded meaning belonging to the name typed or selected by the user is immutable and must never be overwritten by wish randomization. A separate pinned-wish state defaults to that decoded meaning; Refresh replaces only this search-intent pin, while the three unpinned suggestions refresh automatically every 9 seconds and animate independently.
  - **Manual Wish Pinning**: Tapping any unpinned database meaning immediately promotes that exact sentence into the pinned card and makes it the active semantic-search intent. The decoded meaning belonging to the typed/selected name remains unchanged, and the promoted sentence must be removed from the three suggestions to prevent duplication.
  - **Name/Meaning Visual Relationship**: The prototype-name (or surname) input and its decoded database meaning must render as one visually connected group. Label the result with the actual input value (for example `ความหมายของ “ณเดชน์”`) so users cannot mistake a randomized wish for their name's meaning.
  - **Meaning Intent Source**: A pinned DB-derived meaning is sent to semantic ranking as the active meaning itself and is labeled `ความหมาย` in result summaries. The prototype person's name must not be presented or ranked as an “AI Wish”.
  - **Celebrity Avatars Integration**: Upgraded the `GetCelebrities` endpoint to fetch `avatar_url` from the `naming_examples` database table. Added a static file router in the Go backend (`/avatars`) to securely serve the images, and integrated `coil-compose` (`AsyncImage`) in the Android frontend to dynamically render rich circular avatars, falling back to colorful initials if images are unavailable.
  - **Blessing Card Screen Refactoring & UI Polish**: Fixed ViewModel state flow references (`selectedItem`, `surname`), resolved theme color missing definition (`DarkBackground`), updated the Day of Birth Kalakini status UI to render continuous Thai text without brackets or emojis (`เกิดวันอาทิตย์ ปลอดกาลกิณี`), and added an elegant sacred text seal `✦ ดีทั้งเลขศาสตร์และพลังเงา ✦` in warm bronze-gold (`#8B5E14`) tailored for wallpaper capture.
- **4D Ranking Integrity**: Every candidate whose number pairs exist in the governed dictionary remains scoreable. PairPoint contributes through the เลขศาสตร์ factor and non-D pairs carry a caution label; they do not trigger a quota or a weight-dependent hidden exclusion. The initial candidate cache is a bounded, weight-independent superset, so slider reranks remain memory-only with no repeated database or embedding call.
  - **Weight-Conscious Tie-Breaker Logic**: Updated `scoring_filter.go` so fallback tie-breakers strictly check `numWeight > 0` and `semWeight > 0` before using numerology/semantic scores to break ties, ensuring zero-weighted dimensions never secretly influence search order.
  - **Name Length Sound Weighting**: Incorporated UTF-8 rune count into `calculateSoundScore` in `scoring_filter.go`. Names between 5–9 runes suffer 0 penalty, while names shorter than 5 or longer than 9 receive a 2 point/rune penalty on `RawSoundScore` to increase score variance when adjusting the triangle graph.
  - **Supported Combined Directions**: The shared `thname` corpus may supply either a first-name or surname candidate. Combined requests explicitly send `search_target` so API policy/provenance and native UI agree on the candidate role; no duplicate surname table is required.
  - **Responsive Name Scaling**: Updated `BlessingCardScreen.kt` to dynamically scale font size based on `fullName.length`, preventing multi-line text wrapping on long names.
  - **Clean Empty State UI**: Refactored `NameResultsScreen.kt` empty results section to display a simple, styled card stating `"ไม่พบรายชื่อที่ผ่านเกณฑ์ตามเงื่อนไขที่กำหนด"`.
  - **Forced Dark Luxury Theme on Results Screen**: Enforced `MaterialTheme(colorScheme = DarkColorScheme)` on `NameResultsScreen.kt` so the screen maintains its elegant dark purple/gold theme even when the Android OS system theme is set to Light Theme.
  - **Store-verified lifetime purchase**: Android Billing and StoreKit 2 accept only `premium_lifetime`. The native receipt/purchase token is sent to `POST /api/v1/entitlements/verify`; Go verifies it with Google Play Developer API or App Store Server API and returns a short-lived signed entitlement token. Tokens are stored with Android Keystore-backed AES/GCM or iOS Keychain and sent as `X-Entitlement-Token`. A local flag, debug build, acknowledgement, or `ITEM_ALREADY_OWNED` response never unlocks content by itself.
  - **Debug-only entitlement override (Mock purchase)**: For emulator/UI verification, Android and iOS expose a Debug-only `Mock purchase` switch through a bug icon in the app bar. This is a presentation-state override, not a StoreKit/Google Play transaction and not a backend entitlement: it must never create, persist, or send a receipt, purchase token, or entitlement token. The override is compiled/guarded out of Release, TestFlight, and Google Play builds. Keep the real purchase path and server verification unchanged; test normal purchases with the switch off, and test Store/Sandbox receipt-token verification through TestFlight or Google Play testing tracks.
  - **Combined Meaning Logic (First Name + Surname)**: `CombineMeanings` in `domain/name_search.go` joins candidate first-name and surname meanings with a simple space only. Do not insert an extra `และ` between the two meanings; the two phrases read naturally as one story when separated by whitespace. `cachedCandidatePool` and `scoreCandidatePool` preserve `surnameMeaning` across pipeline scoring runs so combined mode candidate items contain their complete joined meaning.
  - **Android network trust**: Retrofit uses the Android platform trust store and normal hostname verification. Release traffic is HTTPS-only, cleartext is disabled, and neither a forced Host header nor trust-all TLS manager is permitted.
  - **Coil Avatar Image Asset Loading Fix**: Initialized Coil's global `ImageLoader` in `MainActivity.kt` with `RetrofitClient.okHttpClient` and updated `MiracleApiConfig.assetUrl` to target domain `http://xn--b3cu8e7ah6h.com/`, enabling circular celebrity avatars to load cleanly over network requests.
  - **Website Download Portal Integration**: Added official Google Play Store link (`https://play.google.com/store/apps/details?id=com.numberniceic.app`) and direct APK download route (`/download/apk`) to `static/download.html` and deployed live to `https://ชื่อดี.com/download`.
  - **2026-08-12 Download Page Alignment**: Updated `/download` to use the public Google Play identity `ชื่อดี`, developer `numberniceic`, Android availability, content rating, in-app purchase disclosure, and the verified Play Store URL. Removed unverified iOS/cloud claims and overpromising update/security copy; the landing page now describes the same core features as the Play listing.
  - **2026-08-12 Download Page Showcase**: Added an optimized four-image gallery using the current Android app screenshots. Added an iOS development announcement marked `กำลังพัฒนา • เตรียมเปิดตัวเร็ว ๆ นี้`; it intentionally has no App Store download link until an official listing exists.
  - **2026-08-12 Web Brand Copy**: Standardized the promotional website identity and calls to action to `ชื่อดี-ตั้งชื่อมงคล`. Updated landing-page SEO metadata and replaced the unverifiable `แม่นยำที่สุด` claim with concrete product capabilities aligned with the Google Play listing.
  - **2026-08-13 Landing Page Dual-Store Integration**: Updated `index.html` and `download.html` in `go-backend` to feature dual store buttons for Android (Google Play) and iOS (App Store) with high-aesthetic dark-emerald gold container styling, removed direct APK download buttons, and deployed live to `https://ชื่อดี.com`.
  - **2026-08-13 iOS App Store Rejection Fix & Upload**: Resolved Apple Guideline 2.1.0 (App Completeness) rejection for submission `1.0.0 (2)`. Updated StoreKit 2 product ID fetching in `PurchaseManager.swift` to search candidate set (`premium_lifetime`, `com.numberniceic.app.premium_lifetime`, `unlock_all`, `com.numberniceic.app.unlock_all`), added purchase alert handling to `BlessingCardView.swift`, added mandatory Terms of Use (EULA) and Privacy Policy links to `RankingResultsView.swift` and `SearchView.swift` (Guidelines 3.1.2 & 5.1.1), bumped `CURRENT_PROJECT_VERSION` to `3`, and successfully archived and uploaded `ChueDee 1.0.0 (3)` to App Store Connect.
  - **2026-08-10 Android Brand/UI Polish**: Adopted the `ชื่อดี` app identity with the selected green-gold `ช` logo for launcher/splash assets. The first-screen mode segmented control uses labels `เฉพาะชื่อตัว` and `ชื่อ + นามสกุล`; combined mode can rank either side from the shared corpus using the search-target toggle. The main action button is intentionally simpler and more premium: gold gradient, `จัดอันดับชื่อดีตามตำรา`, subtle sparkle icons only, no crystal-ball emoji. Ranking tables are full-bleed with no outer left/right row padding. The pair detail sheet uses a dark emerald theme for comfort and shared `PairCircles`.
  - **2026-08-11 Guidance UX Decision**: Do not use a blocking Spotlight Coachmark Tour on the main search screen. Prefer self-explanatory labels, short contextual helper text, sensible defaults, immediate feedback, and empty/error states that explain the next action at the moment it is needed.
  - **2026-08-11 Blessing Card Image Loading UX**: The blessing wallpaper screen should never show an empty Buddha-image area while assets are loading. Use a soft golden circular light placeholder with a gently pulsing halo and small sparkles in the same layout slot, then fade in the real Buddha image when Coil succeeds.
- **Immediate Next Steps**:
  1. Restore a current backup into staging and apply/audit migrations `000019`–`000022`; do not start with production.
  2. Obtain data-owner decisions for the 42 legacy `แ` sum rows, broader title-prefix candidates, and any remaining constraint findings.
  3. Regenerate/link/approve the semantic vector corpus and approve a human-rated sound calibration snapshot. Until then semantic remains fail-closed and sound remains explicitly provisional.
  4. Re-run backend goldens, Android tests/build, and the read-only preflight in staging before a controlled production rollout/restart.

## 7. Core Algorithms & Logic

### Smart Hybrid Ranking Strategy (Name Search)
The backend employs a multi-tiered sorting algorithm (`scoring_filter.go`) to ensure the most auspicious and relevant names bubble to the top.

**Scoring Calculation:**
- `FinalScore = (NumerologyRaw * NumerologyWeight) + (SemanticRaw * SemanticWeight) + (SoundCalibrated * SoundWeight) + (SimilarityRaw * SimilarityWeight)`
- Default balanced formula: **35% เลขศาสตร์, 25% ความหมาย, 20% ไพเราะ, 20% ความคล้าย**. Android and iOS expose one draggable four-axis diamond: top is เลขศาสตร์, left is ความหมาย, right is เอกลักษณ์ (ความคล้าย), and bottom is ไพเราะ. The drag position produces all four weights and always totals 100%; the backend normalizes them again for safety.
- **User-controlled 4D reranking contract (2026-09-15):** Dragging the diamond changes the four requested weights and debounces a `/api/v1/names/rerank` request against the current opaque candidate token. The Go backend is the sole authority for the resulting `final_score`, global sort, and Top N. The returned list is exactly the highest `final_score` rows in that candidate pool; no client may sort it again or reserve rows by pair category. A changed prototype name starts a new search and therefore a new indexed similarity candidate pool; dragging weights only reranks the existing pool.
- The text field accepts only a prototype first name or surname. Android sends it as `prototype_name`; PostgreSQL `pg_trgm` compares that value with candidate names. Its decoded/pinned meaning is sent separately as `wish_query` for semantic-vector ranking, so a meaning sentence is never passed to trigram similarity.
- **Ranking Formula Display Contract:** A dragged selector follows the finger immediately using pending/requested weights and must not be rebound to a delayed API response. While reranking, explicitly label the preview as `กำลังคำนวณสูตรใหม่` and keep result rows non-interactive. When the response arrives, atomically replace the ordering and confirmed `applied_weights`; both clients display the applied formula as the formula used for the current result set, including any redistribution when semantic scoring is unavailable.
- `SoundRaw = (Euphony * 45%) + (PronunciationEase * 35%) + (Rhythm * 20%)`. The backend maps each `phonetic_version` distribution to a common reference before comparison. Live means/sample standard deviations use exact PostgreSQL `numeric` aggregation followed by one `float64` conversion, avoiding nondeterministic parallel floating-point reduction. It prefers an approved immutable snapshot and otherwise labels its process-cached live fallback as provisional.
- If semantic scoring is unavailable, it remains unavailable with score `0`; the backend returns both requested and applied weights and redistributes the unavailable semantic weight across dimensions backed by real data. A fallback must never manufacture a semantic score of `100`.
- External embedding responses are not assumed byte-for-byte deterministic. Each search/rerank response carries the exact query model and SHA-256 query-vector hash used by its cached candidate pool. Semantic scoring prefers an approved model/template/corpus snapshot. Until that governance snapshot is backfilled, the product owner has confirmed the legacy `meaning_vector` corpus was generated with `text-embedding-3-small`, so the backend may use rows with a 1536-dimensional vector as `owner_confirmed_legacy_embedding`.
- When a user adjusts the weights (e.g., 100% Numerology), the `FinalScore` dynamically reflects that preference. Sound quality may order exact score ties but does not alter the selected 100% Numerology score.
- **Deterministic ranking precision — `ranking_precision_12dp_v1`:** normalized numerology, semantic, calibrated sound, and the weighted final score are rounded to 12 decimal places with round-half-away-from-zero before comparison and response. Differences below this declared precision cannot silently outrank the documented text/ID tie-breakers.
- **Complete candidate pool and live reranking:** A prototype-name search retrieves the complete numerology/Kalakini-eligible similarity universe selected by the indexed raw-or-phonetic trigram predicate, with no weight-dependent retrieval limit. The backend calculates every available ranking dimension in that universe, sorts globally, and only then takes the requested Top N. The response includes an opaque `candidate_token`. The backend keeps at most 8 complete pools for 10 minutes and reruns only `ScoringFilter`; it does not repeat database retrieval or embedding generation. Android debounces reranking, commits results and applied metadata atomically, rejects stale search/rerank responses, and rolls pending weights back on failure.
- **Initial-search query performance:** Full-pool SQL must not sort rows because `ScoringFilter` performs the authoritative global ordering in Go. Dictionary coverage uses one query-level governed-character aggregate plus `translate()` instead of a correlated per-character anti-join. Candidate retrieval carries no meaning text; after scoring, `GetMeaningsByNameIDs` fetches meanings only for the returned Top N. Reranking retains the cached scoring dimensions and performs only this small Top-N meaning lookup.
- **Fail-closed corpus eligibility — `title_and_dictionary_coverage_v2`:** Candidate rows with `เด็กหญิง`/`เด็กชาย` administrative prefixes, a character absent from both numerology dictionaries, or a known canonical stored-sum mismatch are excluded before scoring. This applies the same dictionary-coverage rule to imported names that surname/decode requests already use dynamically.
- **Governed Pair Scoring (before scoring):** Every evaluated pair for both **เลขศาสตร์** and **พลังเงา** must exist in the central number dictionary. Known D and R grades remain scoreable at every weight: their dictionary `pairpoint` values affect the numerology factor, and non-D grades are returned as caution metadata. Unknown or malformed pairs always exclude the candidate. No pair rule may add a hidden fifth ranking dimension or make eligibility depend on a slider weight.
- **Top-N Integrity:** After global weighted scoring, the visible Top N is exactly the N highest `final_score` rows. Caution labels and UI filters may help users inspect pair risk, but must never reserve or replace rows in the ranked result set.
- **Combined-Mode Candidate Expansion:** For name+surname searches, the backend derives a bounded weight-independent superset of base `sat_sum` and `sha_sum` values across the real range in `names_miracle`. Every total whose pairs exist in the governed dictionary is cached, and all four dimensions are scored from that pool during each rerank.

**Ranking Priorities (The Tie-breaker System):**
1. **Priority 1: `FinalScore` (Float - Descending)**
   - Numerology, semantic relevance, and name-sound quality are normalized to 0–100 before the user-selected weights are applied.
   - No perfect-pair bonus is added. The score preserves the real pair-point differences and cannot be artificially pushed to the 100-point ceiling. At 100% Numerology, candidates with the same displayed pairs correctly receive the same score.
   - Comparator precision is the published 12-decimal `ranking_precision_12dp_v1` value, not hidden sub-precision floating-point noise.
2. **Priority 2: practical sound quality**
   - Exact score ties are ordered by calibrated overall sound score. Raw model components are not used as cross-version tie-breakers.
3. **Priority 3: normalized factor scores, Thai name text, then `NameID`**
   - Remaining ties use numerology score, then semantic score, before stable Thai text ordering.
   - Name text provides a user-comprehensible stable order. `NameID` remains only a final technical fallback for otherwise identical records.

**UX & Ranking Philosophy:**
- **Intentional Simplicity (No Over-filtering)**: We strictly limit tie-breakers to pair-number quality, hybrid score, and total score. Overcomplicating the sorting logic makes the algorithm opaque and difficult for the user to understand. A broader pool of top-tier names lets people make the final choice based on personal preference and sound.

### Pair Number Decoding Strategy (การถอดรหัสคู่เลข)
- **Sum-Based Pair Generation**: Pair numbers (คู่เลข) shown in the UI are **NOT** derived by pairing the literal numerological values of sequential letters in a name.
- **Decoding Flow**:
  1. Calculate the Total Sum for **Numerology (เลขศาสตร์ / SatSum)** from the canonical character dictionaries; stored sums are audit inputs, not authority when they disagree.
  2. Calculate the Total Sum for **Shadow Power (พลังเงา / ShaSum)** using the same versioned rule. A character mapped in only one system contributes zero to the other.
  3. Format these total sum integers into number-pair lookup keys (e.g., a sum of 6 becomes `["06"]`, a sum of 24 becomes `["24"]`, and a sum of 163 becomes `["16", "63"]`).
  4. Lookup these generated pairs in the `numbers` dictionary to resolve their auspicious type (D10, D8, R5, etc.) and color logic.
- **Display metadata:** Decode responses include each pair's point and `pair_is_good`. D-type rows are auspicious even when their point value is low or negative; R-type and unknown rows are not.
- **UI Application**: The `PairCircles` component in the frontend exclusively renders these sum-based pairs (e.g., 2 circles for most names) rather than letter-by-letter combinations (e.g., 5 circles).

## 8. Data Doctrine, Auditability & Versioning

Ranking values are deterministic heuristic scores under the documented rules; they are not probabilities, scientific guarantees, or predictions about a person's future. A displayed score such as `90.7` means “90.7 points under this formula and data version,” not “90.7% certain to be a correct name.”

### Versioned domain doctrine

- **Distinct pair keys — `pair_dictionary_v2_distinct_digit_keys`:** Pair keys are authored strings. Single-digit keys such as `6` and leading-zero keys such as `06` are distinct domain entries and may intentionally have different meanings, types, and points. Sum decoding uses the leading-zero pair for a single-digit total (`6 -> 06`), while repository storage still keeps `6` and `06` separate. D-type rows are auspicious eligibility rows even when `pairpoint` is low or negative; R-type rows remain ineligible.
- **Thai sum compatibility — `thai_sum_v1_dictionary`:** Each character uses its `sat_nums` and `sha_nums` values independently. Presence in either dictionary is supported and the absent dimension contributes `0`; absence from both is invalid. Current `แ` values are เลขศาสตร์ `2` and พลังเงา `0`. Of 17,370 corpus rows containing `แ`, 17,328 match this rule and 42 legacy `4/12` rows are stored-sum anomalies; runtime ranking quarantines mismatches until an approved corrective migration is applied.
- **Combined scope:** The searched candidate and combined first-name + surname totals must have governed pairs. Their PairPoint values contribute to the numerology factor, while non-D grades are surfaced as cautions. The fixed user-entered part is preserved and is not independently used to reject results.

### 2026-09-15 Production Ranking Integrity Release

- Deployed the 4D ranking integrity backend to Production host `43.228.85.200`; binary SHA-256: `26fe15da0b570e7717d8ecc1208032c3832bb7c1a5b98b41c4e62192926a78a8`.
- Backup immediately before replacement: `/opt/pj-names-miracle/backups/ranking-20260915T010817Z/pj-names-miracle-server`.
- Verified `pj-names-miracle.service` is active and `/api/v1/readiness` returns `X-Ranking-Contract-Version: adaptive_pairs_v1`.
- Production search verified a 100-row result with requested weights `35/25/20/20`, returned applied weights when semantic was unavailable, and policy scope `first_name_all_known_pairs_scored`.

### 2026-09-15 App Store Privacy and Chat Deletion Readiness

- Published App Privacy in App Store Connect for Name, Photos or Videos, Customer Support, Search History, User ID, and Device ID. Each is marked for app functionality, linked to the user, and not used for tracking. Product page preview now reports “Data Linked to You.”
- Deployed the customer chat-delete API and validated the public `/api/v1/readiness` endpoint, missing-token rejection (`401`), and a disposable conversation lifecycle (`201` create, `204` delete, `401` after deletion).
- Updated and verified the production `/privacy` page. It explains deletion in app versions with the menu and email-based requests for older versions.
- Backend binary SHA-256: `858cc6f6f603522c79200fd1a4f1eb811ee5053375784825c93bd6b1c2faf9ed`; rollback binary: `/opt/pj-names-miracle/backups/appstore-delete-fix-20260915T091325Z/pj-names-miracle-server`.
- App Store IPA `1.0.10 (20)` is local at `/tmp/ChueDee-AppStore-1.0.10/exported/ChueDee.ipa`; archive signature verifies and it has not been uploaded or submitted. Release QA and provider-specific data-retention review remain open.
- **กาลกิณี scope — `kalakini_first_name_v1`:** Current `k_<day>` flags and ranking eligibility cover the candidate first name only. The product must not describe the surname or concatenated full name as free of กาลกิณี unless a future, separately versioned full-name rule is implemented.
- **Sound formula — `sound_45_35_20_v1`:** The raw formula remains `Euphony*45% + PronunciationEase*35% + Rhythm*20%`. Rows from different `phonetic_version` values must not be assumed comparable merely because each component is in the nominal 0–100 range. Either rescore the corpus with one approved source version or use an approved calibration snapshot whose mean/stddev and human evaluation are versioned.
- **Corpus title-prefix status — `administrative_child_title_quarantine_v1`:** Runtime ranking excludes the unambiguous administrative prefixes `เด็กหญิง` and `เด็กชาย`; this closes the verified `เด็กหญิงสุรภา` Top-100 contamination. Broader title-like prefixes remain audit candidates until reviewed `name_type`/`review_status` metadata exists, so legitimate names are not removed by an ambiguous text prefix.
- **Semantic corpus provenance — `semantic_vector_snapshot_v1`:** The preferred production state is an approved candidate-vector snapshot matching the query embedding model, input template, vector dimensions, corpus filter/fingerprint, and held-out evaluation version. During the current product stage, the existing `meaning_vector` corpus is accepted as `owner_confirmed_legacy_embedding` because the product owner confirmed it was generated with `text-embedding-3-small`; runtime uses only candidate rows whose stored vector is present and 1536-dimensional.
- **Ranking precision — `ranking_precision_12dp_v1`:** Factor and final scores use round-half-away-from-zero at 12 decimal places before sorting. This version must change if precision or rounding semantics change.

### Governance requirements

- Database schema, lookup data, model/prompt versions, calibration snapshots, and domain-rule changes must be reproducible from versioned migrations or reviewed seed artifacts. A bare `schema_migrations` version in a remote database is not sufficient provenance.
- Each result should be traceable to raw sums, generated pairs, pair types/points, semantic distance/source, query-vector hash, approved candidate-vector snapshot, three raw sound components, normalized factor scores, selected weights, candidate-pool scope, and rule/data/model versions.
- Duplicate exact pair keys, malformed pair keys, sum recomputation mismatches, title-prefix candidates, missing mappings, and mixed phonetic versions are release-audit findings. None may be “fixed” automatically without an explicit data-owner decision. Distinct keys such as `6` and `06` are intentional, and negative D5 points are a deliberate low-score ranking signal.
- Runtime sound calibration prefers the latest immutable approved snapshot batch. If none exists, the server caches one live per-version profile for the process and marks it `live_repository_distribution_provisional` in the API/UI; production release remains blocked until a snapshot and human evaluation are approved.
- Ranking corpus/provenance is process-cached after validation. Ranking tables and dictionaries are operationally immutable while an API instance serves traffic; any governed data write requires taking affected instances out of traffic, regenerating/reapproving evidence where applicable, and restarting every instance before rankings resume.

### Safe database rollout

The non-destructive governance migrations and read-only preflight live under [`go-backend/db/`](go-backend/db/README.md). The required order is:

1. Run `db/audits/preflight_read_only.sql` with a read-only account and archive its output.
2. Restore a current backup into staging; never test governance DDL first on production.
3. Apply the versioned audit views, `NOT VALID` guard constraints, duplicate-tolerant preflight index/function, and nullable semantic-vector snapshot linkage.
4. Review `แ` recomputation, title-prefix, and per-phonetic-version calibration reports with the data/domain owner.
5. Apply corrections through a separate reviewed migration. Do not delete, merge, or relabel domain rows inside an audit migration.
6. Validate constraints and run the manual canonical unique-index step only after every preflight violation is resolved.
7. Run golden standalone/combined/Kalakini/semantic/sound tests in staging and obtain signoff before production rollout.

### Results table discoverability

- The ranking table is intentionally compact, but each entire row is tappable and opens the name's blessing card for saving as an image.
- A lightweight helper line above the table says `แตะชื่อเพื่อดูการ์ดมงคลและบันทึกเป็นรูปภาพ`.
- On the first results display in a screen session, a single Snackbar reinforces the action with `แตะชื่อใดก็ได้เพื่อเปิดการ์ดมงคล`; no persistent coachmark tour is used.

### Lifetime name unlock

- Before a verified lifetime purchase, the server sends three complete preview names, then only a short prefix followed by `***`; protected meanings and ranking reasons are omitted. Client-side masking is presentation defense-in-depth, not the access boundary.
- Android uses Google Play Billing and iOS uses StoreKit 2 with the shared product identifier `premium_lifetime`.
- Only a store-verified, unexpired server entitlement reveals full candidate names. Free users may open, save, and share blessing cards with the candidate portion masked, so they can evaluate the complete feature without exposing paid content. Both platforms provide a restore-purchase action.
- Required production secrets are `ENTITLEMENT_SIGNING_SECRET` (at least 32 bytes), Google service-account configuration (`GOOGLE_PLAY_SERVICE_ACCOUNT_PATH`, `GOOGLE_PLAY_PACKAGE_NAME`), and App Store Server API configuration (`APP_STORE_ISSUER_ID`, `APP_STORE_KEY_ID`, `APP_STORE_PRIVATE_KEY_PATH`, `APP_STORE_BUNDLE_ID`). Secret files never belong in source control.
- Name suggestions from `pg_trgm` are a premium-only feature on both Android and iOS. The server may send free users only a count and up to two green `เลขศาสตร์`/`พลังเงา` pair previews; it must not send suggestion names, IDs, meanings, or selection payloads. A full suggestion response requires a version-2 entitlement minted after live Google Play/App Store verification. Legacy version-1 tokens remain valid for older ranking results but do not unlock name suggestions. If either store verifier is unconfigured, that platform's purchase verification fails closed; a debug mock or plausible token/transaction ID never grants a suggestion entitlement. See `go-backend/docs/purchase-verification-setup.md` for credential setup.
- Production database transport should set `DB_SSLMODE=require` or `verify-full` when PostgreSQL is reached over a network; local development may explicitly use `disable`.

### Pair presentation contract

- Search rows expose `sat_pairs`, `sha_pairs`, `candidate_sat_pairs`, and `candidate_sha_pairs` as `{value,type,is_good}` objects. Native clients use these fields as the color authority and retain the legacy display strings only for backward compatibility. Every visible pair remains rendered through the shared colored-circle component.

### Cross-platform 4D graph and content layout contract

- Android and iOS use the same diamond orientation everywhere: top = `เลขศาสตร์`, left = `ความหมาย`, right = `เอกลักษณ์` (ความคล้าย/identity), and bottom = `ไพเราะ`.
- The graph's weight calculation and its visual spark position must use that same axis mapping. In particular, right and bottom must never be swapped in either the interactive selector or the profile radar.
- Corner icons in every interactive 4D graph remain outside the diamond. Reserve a visual gutter around the plot, while keeping the left/right `ความหมาย` and `เอกลักษณ์` icons close enough to their vertices to read as axis markers rather than detached controls.
### Auspicious calendar and native mobile API suite

- **Clean Architecture & Business Calculation SSOT**: All astrological rules (Northeastern Thai traditional astrology, Taksa 9-step progression, center-cell ๙ logic, 06:00 AM cutoff, scoring, exclusion) remain 100% in Go backend. Android and iOS native clients own UI state and rendering only.
- **Native Mobile Finder (Android 2026-08-26, iOS 2026-08-28)**: Android exposes the finder from a `CalendarMonth` icon in the main and result-screen app bars; iOS exposes its SwiftUI equivalent from the main header. Both native screens follow the emerald/gold hierarchy, load the activity catalog from Go, collect a valid date range of at most 366 days, require birth date/time for personalized Taksa, support practical preferences, and render the API's explainable top-three recommendations. User-facing dates are Thai Buddhist Era while the API receives Gregorian ISO dates. Both clients render the backend-provided animated 8-sector Taksa wheel, center-cell state, all eight roles, and the ศรี/เดช/มนตรี/กาลกิณี highlights. Taksa placement, lunar conversion, scoring, exclusions, and age remain backend-owned. The selected start date controls only the search range; progressed age is always derived from the birth date against the current date in Thailand, and clients never calculate role placement locally.
- **RESTful Endpoints for Web, Android & iOS**:
  1. `GET /api/v1/auspicious/activities` (alias `/api/v1/auspicious-activities`): Returns full activity catalog (icons, categories, descriptions, rule scopes).
  2. `GET /api/v1/auspicious/taksa` & `POST /api/v1/auspicious/taksa`:
     - Params: `birth_date`, `birth_time`, `target_date`.
     - Returns: `is_today_default`, `birth_day_code`, `birth_day_name`, `birth_planet_num`, `birth_time_shifted`, `full_age`, `progressed_age`, `is_center_cell`, `landed_planet_num`, `landed_planet_name`, `wheel_rotation_degrees`, `hero_roles` (sri, dech, montri, kalakini), and `all_roles` (8 positions).
  3. `POST /api/v1/auspicious/search` (alias `/api/v1/auspicious-search`):
     - Request body: `{ activity_code, start_date, end_date, birth_date, birth_time, prefer_weekend, avoid_wan_phra, monk_ceremony }`.
     - Returns: Recommended dates with evaluated scores, rating labels (`ดีมาก`, `ดี`, etc.), lunar labels, and itemized reasons.
  4. `GET /api/v1/auspicious/calendar` (alias `/api/v1/auspicious-calendar`):
     - Params: `year`, `month`.
     - Returns: Daily calendar days mapped to Thai lunar calendar dates and observances (`is_wan_phra`).
- **Taksa Astrological Personalization (หลักทักษาพยากรณ์ & ทักษาจร 9 จังหวะเข้าตากลาง)**:
  - Users provide `birth_date` (YYYY-MM-DD) and `birth_time` (HH:MM) in the calendar UI.
  - **Astrological 06:00 AM Sunrise Cutoff (ตัดวันใหม่ 06:00 น.)**: In Thai astrology, the new day starts at 06:00 AM sunrise. Births between 00:00–05:59 AM automatically roll back to the previous day; Wednesday births between 18:00–05:59 are classified as Wednesday Night / Rahu (`wednesday2`).
  - **Automatic Astrological Age (คำนวณอายุย่างอัตโนมัติ)**: The engine automatically calculates full age and astrological age in progression (`อายุย่าง = อายุเต็ม + 1`), and displays real-time astrological info.
  - **9-Step Progression with Center Cell (ทักษาจร 9 จังหวะเข้าตากลาง ๙)**:
    - Sequence: `อาทิตย์ (1) -> ตากลาง (9/เกตุ) -> จันทร์ (2) -> อังคาร (3) -> พุธ (4) -> เสาร์ (7) -> พฤหัส (5) -> ราหู (8) -> ศุกร์ (6)` (cycle of 9).
    - When landing on **ตากลาง (9)**: Saturday (7) is used as **กาลกิณีจร (-35)**, Thursday (5) is **ศรีจร (+15)**, Sunday (1) is **เดชจร (+12)**, and Monday (2) is **มนตรีจร (+10)**.
    - When landing on standard planets (1, 2, 3, 4, 7, 5, 8, 6): that planet becomes **บริวารจร** and maps the 8 Taksa positions accordingly.
  - **Interactive Animated Taksa Wheel & Streamlined Reference Panel (กงล้อทักษาจร 50/50 Split)**:
    - **Always-On Default Showcase (แสดงผลตลอดเวลาตามพลังวันปัจจุบัน)**: On initial page load (before user inputs birth date), the wheel automatically spins and displays today's planetary energy (`ทักษาพลังประจำวัน`) with clear call-to-action inviting users to enter their birth date.
    - **Seamless Personalization**: Upon entering birth date/time, the wheel smoothly transitions to user's personalized 9-step progressed Taksa with instant spin reaction.
    - **50/50 Layout Split**: Left half features a generous `245px` SVG wheel with strictly horizontal 0° upright typography and center hub (`อายุย่าง XX ปี` / `พลังวันนี้`). Right half features a streamlined, compact 2x2 chip grid (⭐ วันศรี, 👑 วันเดช, 🤝 วันมนตรี, ⚠️ วันกาลี) that matches the height of the wheel without overpowering the page.
    - **Methodology & Audit Link**: Includes persistent note `✦ สำหรับเพื่อตรวจสอบความถูกต้องตามหลักทักษาจร` linking to dedicated [`/taksa-jora-method`](https://xn--b3cu8e7ah6h.com/taksa-jora-method) page explaining 9-step progression and center-cell rules.
    - Expandable accordion for viewing all 8 roles cleanly without clutter.
  - **ทักษากำเนิด (Natal Taksa)**: When `birth_date` is unselected, the engine falls back to universal general auspicious rules.
- Specific activity rules:
  - `name_change`:
    - **Astrological Activation Principle**: In astrology, the auspicious effect of a new name activates the moment the owner **publicly announces or starts using the name** (e.g. on Social Media, changing profiles, or having peers call them) without waiting for district office registration. Thus, weekend/office-hour restrictions are not astrological blockers; users can pick any auspicious day (including Sat/Sun) to activate the name, and visit the district office on a convenient weekday.
    - **Authentic Dictionary Database**: Names recommended by the 4D naming engine are sourced from verified Thai dictionary words with legitimate semantic meanings (not randomly generated letters/vowels), making them easily verifiable and acceptable by civil registry registrars.
    - **Streamlined District Workflow**: Guided by a simplified 3-step article (`/articles/เปลี่ยนชื่อที่อำเภอต้องเตรียมอะไรบ้าง`) focusing on ID card + House registration + 50 THB fee.
  - `business_opening`: Consolidated single work/business activity for starting new jobs, promotions, and opening businesses/launches. Applies `ปุณณดิถี` bonus score (+20) for abundance and new beginning (Thursday on 5, 10, 15 lunar days) along with official civil registry safety guidance.

- **2026-09-14 Admin Name Management & Auto-Meaning Generation (`/admin/names`)**:
  - **Go Backend & Admin Web Interface**: Added administrative name management system accessible via `https://ชื่อดี.com/admin/names` (Basic Auth).
  - **Features**:
    - Add new names with automatic calculation of numerology sum (`เลขศาสตร์`), shadow energy sum (`พลังเงา`), and individual pair validation.
    - Automatic day-of-birth Kalakini calculation across all 8 birth periods.
    - Automated meaning generation pipeline for automated system ingestion and dictionary enrichment.
    - Search and inspect database corpus names directly from the dashboard.
    - Handled search edge cases such as exact name ranking priority (e.g. "เจเน่" surfaced at #1 rank) and dictionary meaning normalization (avoiding generic "ชื่อบุคคล" placeholders when authentic etymological meanings are available).

- **2026-09-14 UX Refinements (Scroll-to-Top & Birthday Picker Stability)**:
  - **Selection Auto-Scroll**: When a user selects any recommended similar name from the candidate list, the viewport automatically scrolls smoothly to the top of the wizard (`scrollToTopTrigger`).
  - **Birthday Selector Stability**: Fixed jitter/viewport shifting when changing day-of-birth in the dropdown menu so the screen remains completely still and stable during selection.
  - **Birthday Kalakini & Name Evaluation Card (Step 2)**: Added dedicated `KalakiniSummaryCard` in Step 2 across Android (Jetpack Compose) and iOS (SwiftUI). Once a birthday is selected, this card focuses directly on the evaluation of the reference name against the chosen birth day without displaying confusing individual consonant/vowel lists. It displays an auspicious green `ปลอดกาลกิณีประจำวัน... 100%` badge with checkmark shield or an amber notice indicating Kalakini presence with reassuring guidance that 100% Kalakini-free names are filtered in subsequent steps.

- **2026-09-14 iOS Debug Mock Purchase & Magic Suggestions Unlock Resolution**:
  - **Background**: Fixed the bug where tapping the bug icon (mock purchase) on iOS left the "ตรวจคู่เลขชื่อที่ใกล้เคียง" (Magic Suggestions) section empty instead of showing unlocked name suggestions with colored `PairCircles`.
  - **3 Root Causes Identified & Resolved**:
    1. **`KeychainHelper.swift` Save/Delete Race**: `save(key:value:)` previously set the `UserDefaults` simulator fallback before calling `delete(key:)`, which immediately cleared the fallback. Fixed by executing `UserDefaults.standard.set` after `delete()`.
    2. **Premature `onChange` Trigger**: `SearchView.swift` was observing `purchaseManager.effectiveIsUnlocked`, which flipped immediately upon toggling before the asynchronous `verifyPurchase` API call completed and stored the entitlement token. Fixed by observing `purchaseManager.isUnlocked`, ensuring the suggestions refresh only fires after the token is verified and persisted.
    3. **Debounce Cancellation Race**: In `SearchViewModel.swift`, the initial `nameChanged()` debounce task cancelled the high-priority immediate `requestNameSuggestions(immediate: true)` initiated by the purchase refresh. Added `suggestionIsImmediate` guard to prevent debounced requests from cancelling an in-flight immediate request.
  - **Debug Purchase Sheet Polish**: Fixed `DebugPurchaseView` RESET button in `PurchaseManager.swift` with proper `dismiss()`, destructive role styling, and toolbar dismiss button.
  - **Verification**: Verified on iPhone 17 simulator via `xcrun simctl` with test query `"จิรายุ"`. Screen renders all 8 unlocked names with correct pair badges (`24/50`, `40/71`, `33/50`, etc.). All temporary debug file logging removed and clean build verified.

- **2026-09-25 Louis Vuitton Brand Theme Unification & Web Backend Production Deployment**:
  - **Context**: Aligned the Web Backend (`go-backend`) public web portal, article reading views, customer support chat, and administrative dashboards with the unified Louis Vuitton luxury dark palette (`#140B07` Deep Espresso, `#26160E` Warm Coffee, `#160C08` Dark Chocolate Card, `#D4AF37` Gold) used by Android and iOS clients.
  - **Key Updates**:
    - `static/styles.css`: Updated root CSS variables (`--bg-base: #140B07`, `--bg-header: #26160E`, `--bg-card: #160C08`, `--bg-card-hover: #2C1A11`), article CTA gradients, header, floating chat CTA, unread badge, and footer.
    - `static/article-download-cta.css`: Updated CTA banners and store buttons.
    - `static/site.webmanifest` & 11 HTML pages: Set `theme-color` meta tags to `#140B07`.
    - Templates (`templates/*.html`): Updated `support_chat_customer.html`, `support_chat_admin.html`, `admin_dashboard.html`, `name_admin.html`, `article_admin.html`, `article_media.html`, `auspicious_admin.html`, and `dimension_admin.html` to eliminate all emerald/green tones in favor of warm espresso brown and gold accents.
  - **Verification & Deployment**:
    - Local Go test suites passed cleanly (`go test ./...`).
    - Cross-compiled Linux AMD64 binary `pj-names-backend-linux`.
    - Synchronized `/static/` assets and updated binary to production server `43.228.85.200` at `/opt/pj-names-miracle/`.
    - Restarted `pj-names-miracle.service` and verified HTTP 200 health check on local upstream `127.0.0.1:18080` and live CDN `https://ชื่อดี.com`.

- **2026-10-11 Growth Optimization, Freemium Unmasking & Celebrity API Fix**:
  - **Context & Diagnosis**: App deployed ~1 month with low conversion/traffic (3 paid users). Root causes identified: 100% masking was too aggressive (users could not experience value before paying), low discoverability/ASO on stores, and lack of organic distribution channels.
  - **Step 1 Implemented & Verified**:
    - **Freemium Unmasking (Top 3 Names Free)**:
      - Backend already had `freeVisibleNameCount = 3`.
      - Android & iOS clients updated to display top 3 ranked names (`rank <= 3`) unmasked without `***` blur.
      - Blessing Card unmasked for top 3 names, allowing free wallpaper generation/preview and download.
    - **Microcopy Polish**:
      - Blessing Card banner pill updated to `"ปลดล็อกเพื่อดู 100 อันดับ"` (Android & iOS).
      - VIP unlock card subtitle updated to `"ปลดล็อกเพื่อดู 100 ชื่อ"` (Android & iOS).
    - **Production VPS Nginx Proxy Fix**:
      - Fixed bug where celebrity names list disappeared on clients due to `/etc/nginx/sites-enabled/xn--b3cu8e7ah6h.com` proxying `/api/v1/` to legacy port 8095. Corrected to `http://pj_names_miracle_backend` (port 18080). Verified `/api/v1/celebrities` returns 200 with all 11 records.
    - **Verification**: Clean builds on both iOS (`xcodebuild` BUILD SUCCEEDED) and Android (`assembleDebug` BUILD SUCCESSFUL).
  - **Step 2 (ASO & Store Presence Optimization) - Completed**:
    - **Google Play Store (Android)**:
      - Updated Short Description to `"วิเคราะห์ชื่อมงคล 4 มิติ ดูฟรี 3 อันดับแรก พร้อมตรวจเลขศาสตร์ พลังเงา และทักษา"`.
      - Rewrote Full Description emphasizing Freemium trial (3 free names + free blessing cards) and keyword density (`ตั้งชื่อลูก`, `เปลี่ยนชื่อ`, `เลขศาสตร์`, `พลังเงา`, `ทักษา`).
      - Bumped Android version to `1.5.0` (versionCode `32`). Built signed `ChueDee-1.5.0-build32.aab` and submitted to Google Play Production track.
    - **Apple App Store (iOS)**:
      - Created version `1.1.0` on App Store Connect.
      - Updated Promotional Text, Description (clean ASCII formatting without disallowed emojis), and high-intent Keywords field (`ชื่อดี,ตั้งชื่อมงคล,ตั้งชื่อลูก,เปลี่ยนชื่อ,ดูดวงชื่อ,เลขศาสตร์,พลังเงา,ทักษา,ฤกษ์คลอด,ฤกษ์มงคล`).
      - Bumped iOS `MARKETING_VERSION` to `1.1.0` and `CURRENT_PROJECT_VERSION` to `22`. Archived, exported, and uploaded build `22` via Xcode Organizer. Submitted for App Store Review.
  - **Step 3 (Organic Marketing & Web-to-App Funnel) - Completed**:
    - **Web-to-App Conversion Funnel**:
      - Updated `article_download_cta.go` to `"วิเคราะห์ชื่อมงคล 4 มิติ — ดูฟรี 3 อันดับแรก"` with CTA copy highlighting free cards and unmasked preview.
      - Updated `download.html` store-facts pill list to highlight `"✨ ทดลองดูผลวิเคราะห์ฟรี 3 อันดับแรก"` and `"🎴 ดาวน์โหลดการ์ดอวยพรมงคลฟรี"`.
      - Cross-compiled Go backend binary, deployed live to Production server `43.228.85.200`, restarted `pj-names-miracle.service`, and verified live HTTP 200 responses on articles and download pages.
    - **Organic Marketing Assets**: Created viral TikTok/Reels/Lemon8 content scripts targeting baby naming, shadow number checking, and aesthetic blessing card wallpapers.

