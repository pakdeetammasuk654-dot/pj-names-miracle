# Database governance and audit rollout

This directory introduces versioned, non-destructive governance around the
existing PostgreSQL data. It does not contain an automatic correction for any
domain value. In particular, it never deletes a corpus row merely because it
resembles a title-prefixed name.

No SQL in this directory was applied to the configured remote database while
these files were created.

## Versioned doctrine

- **Pair keys — `pair_dictionary_v2_distinct_digit_keys`:** a pair key is the
  authored text value. Single-digit keys such as `6` and leading-zero keys such
  as `06` are distinct domain entries, not aliases. Sum decoding still uses the
  leading-zero pair for a single-digit total (`6 -> 06`). Classified D rows are
  eligible even when `pairpoint` is low or negative; classified R rows require
  `pairpoint < 0`, and zero is not a valid classified score.
- **Thai sums — `thai_sum_v1_dictionary`:** each character uses its
  `sat_nums` and `sha_nums` values independently. A character present in
  either table is supported and contributes zero on the missing side; only a
  character absent from both tables is invalid. In the current dictionaries
  `แ` contributes เลขศาสตร์ `2` and พลังเงา `0`. Full-corpus evidence shows
  17,328 of 17,370 rows containing `แ` follow this rule; the remaining 42
  stored rows match a legacy `4/12` pattern and are anomalies to quarantine
  and correct through an approved data migration.
- **Combined mode — `combined_name_and_full_pairs_v2`:** a candidate must have
  all-D pairs for its first-name เลขศาสตร์ and พลังเงา totals **and** all-D
  pairs for both totals after adding the surname. Surname-only pairs are shown
  for explanation but are not a rejection gate.
- **กาลกิณี — `kalakini_first_name_v1`:** the existing `k_<day>` flags and the
  ranking gate cover the candidate first name only. They do not claim that an
  inherited surname or the concatenated full name is free of กาลกิณี letters.
- **Sound — `sound_45_35_20_v1`:** raw sound is
  `euphony*0.45 + pronunciation_ease*0.35 + rhythm*0.20`. Raw scores from
  different `phonetic_version` values are not assumed to share a calibrated
  distribution. Live per-version means and sample standard deviations are
  aggregated as exact PostgreSQL `numeric` values and converted to `float64`
  once; parallel `double precision` aggregation is forbidden because its
  reduction order can change deep tie rankings between identical runs.
- **Semantic vectors — `semantic_vector_snapshot_v1`:** a query vector may be
  compared with `names_miracle.meaning_vector` only when an approved metadata
  row proves the candidate corpus model, input template, dimensions, corpus
  filter/vector fingerprint/membership fingerprint, and held-out evaluation
  version. Vector dimensions alone
  do not prove model compatibility. If that proof is absent, semantic is
  unavailable and its requested weight is redistributed transparently. An
  approved snapshot must cover every runtime-eligible row so missing vectors do
  not introduce a hidden ranking penalty.
- **Title prefixes — `administrative_child_title_quarantine_v1`:** ranking
  excludes the unambiguous administrative prefixes `เด็กหญิง` and
  `เด็กชาย` after an oracle found `เด็กหญิงสุรภา` in a delivered Top
  100. Broader matches such as `นาย`, `นาง`, or `พระ` remain audit-only
  until reviewed `name_type` or `review_status` metadata exists.
- **Runtime corpus — `title_and_dictionary_coverage_v2`:** ranking also rejects
  any candidate containing a character absent from both numerology dictionaries
  and quarantines known stored-sum rows that disagree with the canonical `แ`
  rule. This makes stored imports obey the same fail-closed character rule as
  dynamic surname/decode requests.

Scores produced from these rules are heuristic scores according to the stated
formula. They are not probabilities or empirical guarantees about a person's
future.

The API additionally declares `ranking_precision_12dp_v1`: normalized factor
scores and the weighted final score use round-half-away-from-zero at 12 decimal
places before sorting. This prevents unreported floating-point sub-precision
from deciding deep ties.

## Files

- `audits/preflight_read_only.sql` is self-contained and runs inside a
  read-only transaction. It can be executed before any migration.
- `migrations/000019_*` creates a rule-version registry and lazy audit views.
  It also creates live mean/stddev calibration views for the exact sound
  formula plus empty tables for later approved sound-calibration and semantic-
  vector snapshots.
- `migrations/000020_*` adds `NOT VALID` checks. PostgreSQL enforces them for
  new or updated rows without scanning, rewriting, deleting, or validating
  legacy rows.
- `migrations/000021_*` adds a duplicate-tolerant canonical lookup index and a
  read-only assertion function. It does not create a unique index.
- `migrations/000022_*` adds nullable per-row candidate-vector snapshot linkage,
  live approval checks, and append-only protection for approved semantic/sound
  evidence. Legacy vectors remain unlinked and therefore unavailable.
- `manual/000023_enforce_canonical_pairs_after_cleanup.sql` is deliberately
  outside the automatic migration sequence. It validates preconditions before
  adding canonical uniqueness.

The repository did not previously contain migrations. The configured database
was observed at schema version `18`, so these files begin at `19`; confirm the
target's migration history before adopting these version numbers.

## Read-only preflight

Use a read-only database account where possible. With standard PostgreSQL
connection variables configured:

```sh
cd go-backend
psql -X -v ON_ERROR_STOP=1 -f db/audits/preflight_read_only.sql
```

Review and archive the output. A release is blocked until owners explicitly
review at least:

1. every duplicate exact pair key or malformed pair key;
2. every zero-point D row and R-type non-negative point;
3. canonical dictionary sum deltas, including the 42 observed legacy `แ=4/12`
   anomaly candidates and characters absent from both dictionaries;
4. title-prefix candidates and their corpus disposition;
5. per-`phonetic_version` mean and standard deviation for
   `sound_45_35_20_v1`;
6. unvalidated constraints and any out-of-range score rows.

The preflight intentionally returns findings instead of treating known legacy
anomalies as a SQL error. CI or a release pipeline may compare its aggregate
counts against a reviewed baseline, but a baseline increase must fail review.

## Staged rollout

1. Take a recoverable database backup and restore it into staging.
2. Run `preflight_read_only.sql` against staging before migrations.
3. Apply `000019` through `000022` **up** migrations in order through
   the project's approved migration runner. Do not run down files or the
   `manual/` directory as part of that sequence.
4. Query the installed audit views and rerun the self-contained preflight.
5. Have the domain/data owner approve a separate corrective migration. It must
   state how the 42 stored sums that disagree with `thai_sum_v1_dictionary` are
   corrected. Do not infer pair meanings from grade prefixes or point signs.
6. Confirm the `เด็กหญิง`/`เด็กชาย` runtime quarantine, then resolve
   broader title-prefix candidates through reviewed metadata. Do not silently
   delete or filter ambiguous prefix matches.
7. Regenerate/version the meaning-vector corpus and approve its model, input
   template, dimensions, corpus SHA-256, and held-out evaluation. Until then,
   semantic scoring deliberately remains unavailable.
8. Rescore all phonetic rows with one source version, or approve a calibration
   snapshot. `v_phonetic_calibration_stats` computes live mean/stddev without
   hard-coded production values; `v_phonetic_live_zscores` is exploratory and
   may drift as rows change. Runtime prefers an approved batch in
   `phonetic_calibration_snapshots`; when none exists, the API labels its
   process-cached live fallback as provisional. A production release still
   requires an approved snapshot.
9. After duplicate-key and format reports are empty, run
   `manual/000023_enforce_canonical_pairs_after_cleanup.sql` in staging. This
   calls `assert_canonical_pairs_ready()` before creating the unique index.
10. Validate each remaining `NOT VALID` constraint individually only after its
   violation query is empty. Constraint validation should be scheduled with
   normal production-lock monitoring.
11. Repeat the full search/rerank golden tests in staging, obtain owner signoff,
    and only then repeat the approved steps in production.

The API validates and caches sound/semantic corpus provenance when a process
loads it; it does not continuously poll for database mutations. Treat ranking
tables and dictionaries as immutable for the lifetime of each API process.
After any governed corpus, dictionary, pair, phonetic-score, or vector write,
remove affected instances from traffic, regenerate/reapprove evidence as
required, and restart **every** API instance before serving another ranking.
This restart/revalidation rule is mandatory, not an optional cache hint.

## Phonetic calibration approval

The migration does not embed the currently observed Typhoon means. After a
reviewed audit, capture the exact `formula_version`, `corpus_filter_version`,
`corpus_fingerprint_sha256`, `phonetic_version`, `sample_count`, `sound_mean`,
`sound_stddev`, and audit timestamp in `phonetic_calibration_snapshots`.
Approval also requires a
non-empty held-out `evaluation_version`, `approved_at`, `approved_by`, and
approval note. Changing the source model, prompt, score formula, corpus, or
human evaluation requires a new snapshot; never overwrite an approved snapshot
in place.

Calibration does not establish that the sound labels are correct. Before a
calibrated score affects ranking, validate it against a held-out, human-rated
Thai-name set and publish the evaluation version alongside the snapshot.

For reproducible fingerprints, hash UTF-8 newline-delimited rows in ascending
`name_id` order from `v_runtime_eligible_name_rows`. Each sound row is compact
JSON with keys in this exact order: `name_id`, `thname`, `phonetic_version`,
`euphony_score`, `pronunciation_ease`, `rhythm_score`; blank phonetic versions
become `<missing>` and null components become `0`. Use standard JSON string
escaping with HTML escaping disabled (so `<missing>` remains literal), no
insignificant spaces, and UTF-8 output. Append a newline after every object
(including the last), and hash those bytes. The semantic vector row is
`name_id:meaning_vector::text` plus a
newline. Semantic membership is a separate SHA-256 over `name_id\n` for those
same eligible rows, allowing runtime to recompute this compact proof without
transferring every vector. Record lowercase SHA-256 hex digests.
