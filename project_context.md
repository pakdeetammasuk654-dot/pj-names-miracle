# Project Context: Auspicious Name Analysis & Signature Design

This document serves as the Single Source of Truth (SSOT) for the project's architecture, tech stack, development guidelines, and current status.

## 1. Project Scope
A full-stack application focused on two main features:
- **Auspicious Name Analysis**: Semantic search and analysis utilizing a database of ~300,000 entries.
- **Custom Signature Design**: A feature to generate, design, and manage custom signatures.

## 2. Tech Stack & Core Technologies

### Frontend
- **Platform**: Native Android
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
- `/go-backend/`: Contains the Go application structured following Clean Architecture.

## 5. Development Strategy & Rules

- **UX & Search Mental Model**:
  - Step 1 strictly controls **Calculation Scope** with 2 options: **เฉพาะชื่อตัว (Standalone Name)** vs **รวมชื่อ + นามสกุล (Name + Surname)**.
  - Step 2 Day of Birth selection utilizes an `ExposedDropdownMenuBox` (Dropdown List) to optimize screen space.
  - Global Ranking Weight Ratio (`⭐ เลขศาสตร์ X% 🎯 ความหมาย Y%`) is displayed centrally inside the top pinned sub-header summary card (CardBgDark) to avoid repeating redundant percent text on every item card.
  - Unnecessary "2-syllable filter" UI checkbox removed.
  - All names (1-syllable or 2+ syllables) compete naturally in the ranking based purely on their astrological and semantic scores without artificial rejection.
- **Compose Standards**: All new UI components must be built using Jetpack Compose following modern state hoisting and unidirectional data flow principles. Reusable composables should be prioritized.
- **Terminology Standards**:
  - Main Vowel/Auspicious Number Sum: **เลขศาสตร์**
  - Consonant/Shadow Number Sum: **พลังเงา**
- **UI Component Standards**:
  - All number pairs (คู่เลข) presented across the app (both in search results and Blessing Cards) MUST be rendered as colored circles (`PairCircles`) indicating their auspiciousness level, rather than plain text brackets.
- **Logic Separation**: No business logic or complex data transformations should exist in the Android frontend. If a calculation is needed, it must be migrated to the Go backend and exposed via API.

## 6. Current Project Status & Immediate Next Steps

- **Current Status**: Project skeleton and core search/decoding architecture are established. The workspace contains the `android-frontend` (Kotlin/Compose) and `go-backend` (Go Clean Architecture) connected to PostgreSQL.
- **Recent Technical Implementations**:
  - **Dynamic Sample Names (Celebrities)**: `GetCelebrities` endpoint in Go backend now queries the `sample_names` database table dynamically (`SELECT name FROM sample_names`) via Clean Architecture repository/usecase layers instead of hardcoded lists.
  - **Full-Bleed Celebrity Marquee Layout**: Refactored `SearchWizardScreen` layout hierarchy by scoping horizontal padding to inner child cards rather than the parent scrollable column, allowing the celebrity `LazyRow` to span full screen width seamlessly without edge gaps.
  - **Single-Tap Gesture & Smooth Marquee**: Replaced continuous `animateScrollBy` with frame-based `dispatchRawDelta` via `withFrameNanos` to prevent `isScrollInProgress` from locking tap gestures. Avatar items now respond to a **single tap** to stop scrolling and trigger immediate name decoding.
  - **Wish Auto-Selection Logic**: Manual meaning input has been removed. If a searched name does not exist in the database (no meaning), the system auto-selects the first preset wish (`🌸 เสริมดวงมงคลรอบด้าน (ทุกมิติชีวิต)`) as default to seamlessly enable hybrid/semantic ranking.
  - **Search Mode UI Animations**: Implemented `AnimatedVisibility` (expandVertically + fadeIn) for mode toggles and analysis sections to ensure smooth, natural UX transitions when switching between Standalone and Combined name search modes.
  - **Dynamic Wish Pinning & Auto-Refresh**: Refactored the wish list logic using `LaunchedEffect` to automatically pin both the user's selected meaning and the DB-derived name meaning (💡) to the top of the list. The auto-shuffle interval (9 seconds) strictly respects these pinned items, only shuffling unpinned meanings.
  - **Celebrity Avatars Integration**: Upgraded the `GetCelebrities` endpoint to fetch `avatar_url` from the `naming_examples` database table. Added a static file router in the Go backend (`/avatars`) to securely serve the images, and integrated `coil-compose` (`AsyncImage`) in the Android frontend to dynamically render rich circular avatars, falling back to colorful initials if images are unavailable.
  - **Blessing Card Screen Refactoring & UI Polish**: Fixed ViewModel state flow references (`selectedItem`, `surname`), resolved theme color missing definition (`DarkBackground`), and updated the Day of Birth Kalakini status UI to render continuous Thai text without brackets or emojis (`เกิดวันอาทิตย์ ปลอดกาลกิณี`).
- **Immediate Next Steps**:
  1. Continue refining search result filtering and Blessing Card detail views.
  2. Setup and test edge cases for combined name + surname calculations with surname database entries.
  3. Prepare production builds for Android frontend and Go backend.

## 7. Core Algorithms & Logic

### Smart Hybrid Ranking Strategy (Name Search)
The backend employs a multi-tiered sorting algorithm (`scoring_filter.go`) to ensure the most auspicious and relevant names bubble to the top.

**Scoring Calculation:**
- `FinalScore = (NumerologyRaw * NumerologyWeight) + (SemanticRaw * SemanticWeight)`
- When a user adjusts the weight (e.g., 100% Numerology), the `FinalScore` dynamically reflects that preference.
- **Auspicious Pair Safety Gate (before scoring):** Every generated pair for both **เลขศาสตร์** and **พลังเงา** must exist in the central number dictionary and be a D-type pair. A candidate containing any R-type or unknown pair is excluded completely, including pairs created only after combining the name with a surname. Semantic relevance can never compensate for a bad pair.
- **Combined-Mode Candidate Expansion:** For name+surname searches, the backend derives all base `sat_sum` and `sha_sum` values across the real range in `names_miracle` whose post-surname totals contain only D-type pairs. PostgreSQL filters the full name corpus by these eligible values before semantic ranking, avoiding the old behavior where a small preselected candidate set collapsed to only a few results after surname validation.

**Ranking Priorities (The Tie-breaker System):**
1. **Priority 1: `FinalScore` (Float - Descending)**
   - Both numerology and semantic relevance are normalized to 0–100 before the user-selected weights are applied.
   - A name with perfect D-type pairs for both Numerology and Shadow Power receives a visible +8 bonus, rather than being hard-pinned above the user’s selected semantic intent.
2. **Priority 2: `TotalScore` (Integer - Descending)**
   - The raw numerology score (total points from pair meanings). Used as a tie-breaker if `FinalScore` is identical.
3. **Priority 3: `NameID` (Ascending)**
   - A stable final tie-breaker prevents results with equal scores from reshuffling between identical requests.

**UX & Ranking Philosophy:**
- **Intentional Simplicity (No Over-filtering)**: We strictly limit tie-breakers to pair-number quality, hybrid score, and total score. Overcomplicating the sorting logic makes the algorithm opaque and difficult for the user to understand. A broader pool of top-tier names lets people make the final choice based on personal preference and sound.

### Pair Number Decoding Strategy (การถอดรหัสคู่เลข)
- **Sum-Based Pair Generation**: Pair numbers (คู่เลข) shown in the UI are **NOT** derived by pairing the literal numerological values of sequential letters in a name.
- **Decoding Flow**:
  1. Calculate (or retrieve from DB) the Total Sum for **Numerology (เลขศาสตร์ / SatSum)**.
  2. Calculate (or retrieve from DB) the Total Sum for **Shadow Power (พลังเงา / ShaSum)**.
  3. Format these total sum integers into 2-digit pairs (e.g., a sum of 24 becomes `["24"]`, a sum of 163 becomes `["16", "63"]`).
  4. Lookup these generated pairs in the `numbers` dictionary to resolve their auspicious type (D10, D8, R5, etc.) and color logic.
- **UI Application**: The `PairCircles` component in the frontend exclusively renders these sum-based pairs (e.g., 2 circles for most names) rather than letter-by-letter combinations (e.g., 5 circles).
