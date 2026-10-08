# Name import pipeline

`name-import` is the safe entry point for a new name corpus. It intentionally
does **not** insert directly into `public.names_miracle`: rows first enter the
staging tables created by migration `000024_name_import_staging`.

## First 20-name trial

1. Create a UTF-8 text file with one Thai given name per line. A CSV is also
   accepted; its first column is used as the name.
2. Run preflight. This has no database side effects:

   ```sh
   cd go-backend
   go run ./cmd/name-import -mode preflight -input /absolute/path/names-20.txt
   ```

3. Review the JSON report. It reports blank, duplicate, non-Thai and
   unsupported-character entries before any database write.
4. On a staging database only, apply migration `000024` through the project's
   approved migration runner, then stage the reviewed input:

   ```sh
   go run ./cmd/name-import \
     -mode stage \
     -input /absolute/path/names-20.txt \
     -batch names_aug_2026_trial_20 \
     -created-by tayap \
     -apply
   ```

The command loads database configuration from the same untracked `.env` used
by the backend. It never accepts credentials on its command line.

## Required enrichment before promotion

Promotion is deliberately a later, reviewed operation. A staged row must have:

1. Canonical `sat_sum` and `sha_sum`, calculated from the live `sat_nums` and
   `sha_nums` dictionaries. A character absent from both dictionaries is a
   rejection, not a zero-value guess.
2. The eight first-name `k_*` กาลกิณี flags calculated from `kakis_day`.
3. AI-reviewed `meaning` and `gender` (`ช`, `ญ`, or `ค`), with the exact model
   and prompt version recorded.
4. `euphony_score`, `pronunciation_ease`, `rhythm_score`, and a single declared
   `phonetic_version`.
5. A 1536-dimensional `meaning_vector` created with the runtime input template
   `ความหมาย: <meaning>` and `text-embedding-3-small`.
6. A semantic snapshot and phonetic calibration review that cover the expanded
   eligible corpus, followed by an API restart after the governed write.

The older scripts used an embedding input containing both name and meaning.
Do not reuse that format: it differs from the current runtime contract.
