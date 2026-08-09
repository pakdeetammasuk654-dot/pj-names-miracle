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
  - Global Ranking Weight Ratio (`⭐ เลขศาสตร์ X% 🎯 ความหมาย Y% 🎶 เสียงชื่อ Z%`) is displayed once in the compact pinned results header, not repeated on every ranking row.
  - Ranking results should use a compact full-width table layout, not large cards. Each row has no rounded corners, uses thin row separators, and is divided into columns for rank, name, auspicious number pairs, score, and navigation affordance. Number pairs must be their own column. Include a sticky column header row (`#`, `ชื่อ`, `คู่เลข`, `คะแนน`) so users understand the table while scrolling. The rightmost action affordance should be an icon-only chevron in its own trailing space; avoid repeated action text such as `ดู` in every row because it reads like table data. The whole row remains clickable. Do not show meaning text in the row; users are scanning names first and can open details when needed.
  - Unnecessary "2-syllable filter" UI checkbox removed.
  - All names (1-syllable or 2+ syllables) compete naturally in the ranking based purely on their astrological and semantic scores without artificial rejection.
- **Compose Standards**: All new UI components must be built using Jetpack Compose following modern state hoisting and unidirectional data flow principles. Reusable composables should be prioritized.
- **Typography**: Android uses Sarabun as the primary app font through the global Compose `MaterialTheme` typography. Number-pair circles (`PairCircles`) intentionally keep `FontFamily.Default` so numerology digits remain compact and visually stable.
- **Thai Buddha Assets**: Blessing cards use Photoshop-style transparent PNG assets from the backend static route `/buddhas/{day}_buddha.png`, then fall back to bundled Android `drawable-nodpi` PNG resources when the dev backend is offline. Do not fall back to SVG for this screen; PNG is the visual direction because it is easier to compose as a top-right wallpaper ornament. During development these Coil requests disable memory and disk cache, while keeping network fetching enabled and adding a timestamp query string so replacing image files in `go-backend/thai_buddha/` is reflected after the Android screen reloads. Do not disable Coil network cache policy here, because that can prevent remote images from loading. When adding the remaining weekday Buddha images, add backend PNG files first, then matching Android fallback PNG resources.
- **Wednesday Buddha Naming**: Use `wednesday1_buddha.png` for Wednesday daytime and `wednesday2_buddha.png` for Wednesday night. Android day keys are `wednesday1` and `wednesday2`; keep backend PNG filenames and Android `drawable-nodpi` fallback resource names aligned with those keys.
- **Blessing Card Visual Direction**: The blessing card is intended to be captured and reused as a phone wallpaper. Use a gold-card visual direction for the captured card itself: warm gold gradient background, dark purple/brown text for legibility, Buddha artwork placed at the top-right as a decorative watermark/ornament, centered narrow text column, and numerology meanings below their pair circles.
- **Terminology Standards**:
  - Main Vowel/Auspicious Number Sum: **เลขศาสตร์**
  - Consonant/Shadow Number Sum: **พลังเงา**
- **UI Component Standards**:
  - All number pairs (คู่เลข) presented across the app (both in search results and Blessing Cards) MUST be rendered as colored circles (`PairCircles`) indicating their auspiciousness level, rather than plain text brackets.
- **Logic Separation**: No business logic or complex data transformations should exist in the Android frontend. If a calculation is needed, it must be migrated to the Go backend and exposed via API.

## 6. Current Project Status & Immediate Next Steps

- **Current Status**: Exhaustive, auditable ranking is implemented end to end. Backend golden tests are independently checked against read-only SQL oracles; Android freezes the applied search/rerank formula and provenance with each result set. Database governance SQL is staged locally but has not been applied to production.
- **Recent Technical Implementations**:
  - **Dynamic Sample Names (Celebrities)**: `GetCelebrities` endpoint in Go backend now queries the `sample_names` database table dynamically (`SELECT name FROM sample_names`) via Clean Architecture repository/usecase layers instead of hardcoded lists.
  - **Full-Bleed Celebrity Marquee Layout**: Refactored `SearchWizardScreen` layout hierarchy by scoping horizontal padding to inner child cards rather than the parent scrollable column, allowing the celebrity `LazyRow` to span full screen width seamlessly without edge gaps.
  - **Single-Tap Gesture & Smooth Marquee**: Replaced continuous `animateScrollBy` with frame-based `dispatchRawDelta` via `withFrameNanos` to prevent `isScrollInProgress` from locking tap gestures. Avatar items now respond to a **single tap** to stop scrolling and trigger immediate name decoding.
  - **Wish Auto-Selection Logic**: Manual meaning input has been removed. If a searched name does not exist in the database (no meaning), the system auto-selects the first preset wish (`🌸 เสริมดวงมงคลรอบด้าน (ทุกมิติชีวิต)`) as default to seamlessly enable hybrid/semantic ranking.
  - **Search Mode UI Animations**: Implemented `AnimatedVisibility` (expandVertically + fadeIn) for mode toggles and analysis sections to ensure smooth, natural UX transitions when switching between Standalone and Combined name search modes.
  - **Dynamic Wish Pinning & Auto-Refresh**: Refactored the wish list logic using `LaunchedEffect` to automatically pin both the user's selected meaning and the DB-derived name meaning (💡) to the top of the list. The auto-shuffle interval (9 seconds) strictly respects these pinned items, only shuffling unpinned meanings.
  - **Meaning Intent Source**: A pinned DB-derived meaning is sent to semantic ranking as the active meaning itself and is labeled `ความหมาย` in result summaries. The prototype person's name must not be presented or ranked as an “AI Wish”.
  - **Celebrity Avatars Integration**: Upgraded the `GetCelebrities` endpoint to fetch `avatar_url` from the `naming_examples` database table. Added a static file router in the Go backend (`/avatars`) to securely serve the images, and integrated `coil-compose` (`AsyncImage`) in the Android frontend to dynamically render rich circular avatars, falling back to colorful initials if images are unavailable.
  - **Blessing Card Screen Refactoring & UI Polish**: Fixed ViewModel state flow references (`selectedItem`, `surname`), resolved theme color missing definition (`DarkBackground`), updated the Day of Birth Kalakini status UI to render continuous Thai text without brackets or emojis (`เกิดวันอาทิตย์ ปลอดกาลกิณี`), and added an elegant sacred text seal `✦ ดีทั้งเลขศาสตร์และพลังเงา ✦` in warm bronze-gold (`#8B5E14`) tailored for wallpaper capture.
  - **Strict Auspicious Gate & High Performance**: Enforced strict D-type auspicious pair gating (`allPairsAreGood`) across all search modes to maintain high database query performance and guarantee 100% safety for all returned candidates.
  - **Weight-Conscious Tie-Breaker Logic**: Updated `scoring_filter.go` so fallback tie-breakers strictly check `numWeight > 0` and `semWeight > 0` before using numerology/semantic scores to break ties, ensuring zero-weighted dimensions never secretly influence search order.
  - **Name Length Sound Weighting**: Incorporated UTF-8 rune count into `calculateSoundScore` in `scoring_filter.go`. Names between 5–9 runes suffer 0 penalty, while names shorter than 5 or longer than 9 receive a 2 point/rune penalty on `RawSoundScore` to increase score variance when adjusting the triangle graph.
  - **Dynamic Surname & Mode Headers**: Standardized mode summaries (`โหมด: หาชื่อ + [สกุล]` vs `โหมด: [ชื่อ] + หาสกุล`), toggle labels (`เปลี่ยนเป็น หาสกุล + ชื่อ`), pair label (`คู่เลขศาสตร์ และพลังเงา:` in `SearchWizardScreen.kt`), and dynamic result table headers (`นามสกุล` when in surname-search mode).
  - **Responsive Name Scaling**: Updated `BlessingCardScreen.kt` to dynamically scale font size based on `fullName.length`, preventing multi-line text wrapping on long names.
  - **Clean Empty State UI**: Refactored `NameResultsScreen.kt` empty results section to display a simple, styled card stating `"ไม่พบรายชื่อที่ผ่านเกณฑ์ตามเงื่อนไขที่กำหนด"`.
  - **Forced Dark Luxury Theme on Results Screen**: Enforced `MaterialTheme(colorScheme = DarkColorScheme)` on `NameResultsScreen.kt` so the screen maintains its elegant dark purple/gold theme even when the Android OS system theme is set to Light Theme.
  - **Trigram Similarity (`pg_trgm`) Intent**: Seed name searches (`IsNameIntent: true`) use PostgreSQL `pg_trgm`'s `similarity(thname, seedName)` query, which maps directly into the `SemanticScore` (0–100) range for seed-based similarity ranking.
- **Immediate Next Steps**:
  1. Restore a current backup into staging and apply/audit migrations `000019`–`000022`; do not start with production.
  2. Obtain data-owner decisions for the 42 legacy `แ` sum rows, broader title-prefix candidates, and any remaining constraint findings.
  3. Regenerate/link/approve the semantic vector corpus and approve a human-rated sound calibration snapshot. Until then semantic remains fail-closed and sound remains explicitly provisional.
  4. Re-run backend goldens, Android tests/build, and the read-only preflight in staging before a controlled production rollout/restart.

## 7. Core Algorithms & Logic

### Smart Hybrid Ranking Strategy (Name Search)
The backend employs a multi-tiered sorting algorithm (`scoring_filter.go`) to ensure the most auspicious and relevant names bubble to the top.

**Scoring Calculation:**
- `FinalScore = (NumerologyRaw * NumerologyWeight) + (SemanticRaw * SemanticWeight) + (SoundCalibrated * SoundWeight)`
- Default balanced formula: **45% เลขศาสตร์, 35% meaning, 20% name sound**. Android exposes one draggable ternary triangle: the top vertex represents เลขศาสตร์, bottom-left meaning, and bottom-right name sound. The drag position produces all three weights and always totals 100%; the backend normalizes them again for safety.
- `SoundRaw = (Euphony * 45%) + (PronunciationEase * 35%) + (Rhythm * 20%)`. The backend maps each `phonetic_version` distribution to a common reference before comparison. Live means/sample standard deviations use exact PostgreSQL `numeric` aggregation followed by one `float64` conversion, avoiding nondeterministic parallel floating-point reduction. It prefers an approved immutable snapshot and otherwise labels its process-cached live fallback as provisional.
- If semantic scoring is unavailable, it remains unavailable with score `0`; the backend returns both requested and applied weights and redistributes the unavailable semantic weight across dimensions backed by real data. A fallback must never manufacture a semantic score of `100`.
- External embedding responses are not assumed byte-for-byte deterministic. Each search/rerank response carries the exact query model and SHA-256 query-vector hash used by its cached candidate pool. Semantic scoring prefers an approved model/template/corpus snapshot. Until that governance snapshot is backfilled, the product owner has confirmed the legacy `meaning_vector` corpus was generated with `text-embedding-3-small`, so the backend may use rows with a 1536-dimensional vector as `owner_confirmed_legacy_embedding`.
- When a user adjusts the weights (e.g., 100% Numerology), the `FinalScore` dynamically reflects that preference. Sound quality may order exact score ties but does not alter the selected 100% Numerology score.
- **Deterministic ranking precision — `ranking_precision_12dp_v1`:** normalized numerology, semantic, calibrated sound, and the weighted final score are rounded to 12 decimal places with round-half-away-from-zero before comparison and response. Differences below this declared precision cannot silently outrank the documented text/ID tie-breakers.
- **Complete candidate pool and live reranking:** Initial search retrieves the complete numerology/Kalakini-eligible corpus with no weight-dependent retrieval limit, calculates every available ranking dimension, sorts globally, and only then takes the requested Top N. The response includes an opaque `candidate_token`. The backend keeps at most 8 complete pools for 10 minutes and reruns only `ScoringFilter`; it does not repeat database retrieval or embedding generation. Android debounces reranking, commits results and applied metadata atomically, rejects stale search/rerank responses, and rolls pending weights back on failure.
- **Fail-closed corpus eligibility — `title_and_dictionary_coverage_v2`:** Candidate rows with `เด็กหญิง`/`เด็กชาย` administrative prefixes, a character absent from both numerology dictionaries, or a known canonical stored-sum mismatch are excluded before scoring. This applies the same dictionary-coverage rule to imported names that surname/decode requests already use dynamically.
- **Auspicious Pair Safety Gate (before scoring):** Every evaluated pair for both **เลขศาสตร์** and **พลังเงา** must exist in the central number dictionary and be D-type. Standalone mode evaluates the first-name totals. Combined policy `combined_name_and_full_pairs_v2` requires both the first-name pairs and the post-surname full-name pairs to satisfy that gate; surname-only pairs are explanatory and are not a rejection gate. D5 pairs may carry low or negative `pairpoint` values; they remain eligible but rank lower. Any R-type or unknown pair in an evaluated scope excludes the candidate, and semantic relevance can never compensate for it.
- **Combined-Mode Candidate Expansion:** For name+surname searches, the backend derives all base `sat_sum` and `sha_sum` values across the real range in `names_miracle` whose first-name and post-surname totals both contain only allowed D-type pairs. PostgreSQL filters the full corpus by these values before global scoring.

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
- **Combined scope — `combined_name_and_full_pairs_v2`:** A combined candidate must have all-D first-name เลขศาสตร์/พลังเงา pairs and all-D post-surname full-name pairs. Surname-only pairs remain visible as explanatory context but do not reject a candidate.
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
