-- Data-governance objects only. This migration does not update or delete any
-- production name, number, numerology, Kalakini, meaning, or phonetic row.
--
-- The views intentionally calculate from live source rows. They are audit
-- surfaces, not a silent replacement for the ranking queries.

CREATE SCHEMA IF NOT EXISTS data_governance;

CREATE TABLE IF NOT EXISTS data_governance.rule_versions (
    rule_key text PRIMARY KEY,
    rule_version text NOT NULL,
    status text NOT NULL CHECK (status IN ('documented', 'audit_only', 'approved', 'retired')),
    scope_note text NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO data_governance.rule_versions (rule_key, rule_version, status, scope_note)
VALUES
    (
        'pair_dictionary',
        'pair_dictionary_v2_distinct_digit_keys',
        'audit_only',
        'Pair keys are authored strings. Single-digit keys such as 6 and leading-zero keys such as 06 are distinct entries; D rows are eligible even with low or negative points, R requires pairpoint < 0, and zero is invalid.'
    ),
    (
        'numerology_sum',
        'thai_sum_v1_dictionary',
        'audit_only',
        'Each character uses its sat_nums and sha_nums row independently; a missing side contributes zero, while a character absent from both dictionaries is invalid. Current แ values are เลขศาสตร์ 2 and พลังเงา 0.'
    ),
    (
        'kalakini_scope',
        'kalakini_first_name_v1',
        'audit_only',
        'Current k_<day> flags cover candidate first names only; surname/full-name safety is not implied.'
    ),
    (
        'combined_scope',
        'combined_name_and_full_pairs_v2',
        'approved',
        'Combined candidates must have all-D first-name pairs and all-D post-surname full-name pairs; surname-only pairs are explanatory and not a gate.'
    ),
    (
        'sound_formula',
        'sound_45_35_20_v1',
        'audit_only',
        'Raw sound = euphony*0.45 + pronunciation_ease*0.35 + rhythm*0.20; mixed source versions require calibration.'
    ),
    (
        'semantic_vector_provenance',
        'semantic_vector_snapshot_v1',
        'approved',
        'Semantic scoring fails closed unless the candidate-vector corpus has an approved model, input-template, dimension, corpus fingerprint, and evaluation record.'
    ),
    (
        'corpus_title_prefix',
        'administrative_child_title_quarantine_v1',
        'approved',
        'Ranking excludes unambiguous administrative prefixes เด็กหญิง and เด็กชาย. Broader title-like prefixes remain audit-only until reviewed metadata exists.'
    ),
    (
        'corpus_eligibility',
        'title_and_dictionary_coverage_v2',
        'approved',
        'Ranking excludes unambiguous child-title prefixes, characters absent from both numerology dictionaries, and known stored-sum rows that disagree with the canonical แ rule.'
    )
ON CONFLICT (rule_key) DO NOTHING;

-- Keep the approved values separate from the live statistics view below.
-- A controlled rollout may snapshot reviewed means/stddevs here. This
-- migration deliberately does not hard-code production statistics.
CREATE TABLE IF NOT EXISTS data_governance.phonetic_calibration_snapshots (
    calibration_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    formula_version text NOT NULL,
    corpus_filter_version text NOT NULL CHECK (btrim(corpus_filter_version) <> ''),
    corpus_fingerprint_sha256 text NOT NULL CHECK (
        corpus_fingerprint_sha256 ~ '^[0-9a-f]{64}$'
    ),
    phonetic_version text NOT NULL,
    sample_count bigint NOT NULL CHECK (sample_count > 1),
    sound_mean double precision NOT NULL,
    sound_stddev double precision NOT NULL CHECK (sound_stddev > 0),
    source_audit_at timestamptz NOT NULL,
    evaluation_version text CHECK (
        evaluation_version IS NULL OR btrim(evaluation_version) <> ''
    ),
    approved_at timestamptz,
    approved_by text,
    approval_note text,
    CHECK (
        approved_at IS NULL
        OR (
            btrim(COALESCE(evaluation_version, '')) <> ''
            AND btrim(COALESCE(approved_by, '')) <> ''
            AND btrim(COALESCE(approval_note, '')) <> ''
        )
    ),
    UNIQUE (formula_version, corpus_filter_version, phonetic_version, source_audit_at)
);

-- This table records proof for the vectors already stored on names_miracle;
-- it does not generate, replace, or approve those vectors by itself. Runtime
-- semantic scoring is allowed only when a matching approved row exists.
CREATE TABLE IF NOT EXISTS data_governance.semantic_vector_snapshots (
    snapshot_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    embedding_model text NOT NULL CHECK (btrim(embedding_model) <> ''),
    input_template_version text NOT NULL CHECK (btrim(input_template_version) <> ''),
    vector_dimensions integer NOT NULL CHECK (vector_dimensions > 0),
    corpus_filter_version text NOT NULL CHECK (btrim(corpus_filter_version) <> ''),
    corpus_row_count bigint NOT NULL CHECK (corpus_row_count > 0),
    vectorized_row_count bigint NOT NULL CHECK (
        vectorized_row_count > 0 AND vectorized_row_count <= corpus_row_count
    ),
    corpus_vector_sha256 text NOT NULL CHECK (
        corpus_vector_sha256 ~ '^[0-9a-f]{64}$'
    ),
    corpus_membership_sha256 text NOT NULL CHECK (
        corpus_membership_sha256 ~ '^[0-9a-f]{64}$'
    ),
    source_audit_at timestamptz NOT NULL,
    evaluation_version text CHECK (
        evaluation_version IS NULL OR btrim(evaluation_version) <> ''
    ),
    approved_at timestamptz,
    approved_by text,
    approval_note text,
    CHECK (
        approved_at IS NULL
        OR (
            vectorized_row_count = corpus_row_count
            AND btrim(COALESCE(evaluation_version, '')) <> ''
            AND btrim(COALESCE(approved_by, '')) <> ''
            AND btrim(COALESCE(approval_note, '')) <> ''
        )
    ),
    UNIQUE (
        embedding_model,
        input_template_version,
        corpus_filter_version,
        source_audit_at
    )
);

CREATE OR REPLACE VIEW data_governance.v_canonical_number_rows AS
SELECT
    n.pairnumberid,
    n.pairnumber AS raw_pairnumber,
    btrim(n.pairnumber) AS trimmed_pairnumber,
    CASE
        WHEN btrim(n.pairnumber) ~ '^[0-9]{1,2}$'
        THEN btrim(n.pairnumber)
        ELSE NULL
    END AS canonical_pair,
    btrim(n.pairtype) AS pairtype,
    n.pairpoint,
    n.miracledesc,
    n.miracledetail
FROM public.numbers AS n;

CREATE OR REPLACE VIEW data_governance.v_pair_canonical_collisions AS
SELECT
    canonical_pair,
    count(*) AS row_count,
    count(DISTINCT pairtype) AS distinct_pairtypes,
    count(DISTINCT pairpoint) AS distinct_pairpoints,
    jsonb_agg(
        jsonb_build_object(
            'pairnumberid', pairnumberid,
            'raw_pairnumber', raw_pairnumber,
            'pairtype', pairtype,
            'pairpoint', pairpoint
        )
        ORDER BY trimmed_pairnumber
    ) AS rows
FROM data_governance.v_canonical_number_rows
WHERE canonical_pair IS NOT NULL
GROUP BY canonical_pair
HAVING count(*) > 1;

CREATE OR REPLACE VIEW data_governance.v_pair_grade_point_anomalies AS
SELECT
    pairnumberid,
    raw_pairnumber,
    canonical_pair,
    pairtype,
    pairpoint,
    anomaly_code
FROM data_governance.v_canonical_number_rows
CROSS JOIN LATERAL (
    VALUES
        (CASE WHEN pairtype LIKE 'D%' AND pairpoint = 0 THEN 'zero_d_point' END),
        (CASE WHEN pairtype LIKE 'R%' AND pairpoint >= 0 THEN 'non_negative_r_point' END),
        (CASE WHEN canonical_pair IS NULL THEN 'invalid_pair_key' END)
) AS flags(anomaly_code)
WHERE anomaly_code IS NOT NULL;

CREATE OR REPLACE VIEW data_governance.v_character_dictionary_coverage AS
SELECT
    COALESCE(sat.char_key, sha.char_key) AS char_key,
    sat.sat_value,
    sha.sha_value,
    CASE
        WHEN sat.char_key IS NOT NULL AND sha.char_key IS NOT NULL THEN 'both'
        WHEN sat.char_key IS NOT NULL THEN 'sat_only'
        ELSE 'sha_only'
    END AS mapping_scope
FROM public.sat_nums AS sat
FULL OUTER JOIN public.sha_nums AS sha ON sha.char_key = sat.char_key;

-- SQL mirror of the backend's versioned, day-independent corpus gate. The
-- auspicious-pair and Kalakini filters are request-specific and intentionally
-- applied later.
CREATE OR REPLACE VIEW data_governance.v_runtime_eligible_name_rows AS
SELECT n.*
FROM public.names_miracle AS n
WHERE n.thname NOT LIKE 'เด็กหญิง%'
  AND n.thname NOT LIKE 'เด็กชาย%'
  AND CASE
      WHEN strpos(n.thname, 'แ') = 0 THEN true
      ELSE
          n.sat_sum = COALESCE((
              SELECT sum(COALESCE(s.sat_value, 0))
              FROM regexp_split_to_table(n.thname, '') AS chars(char_key)
              LEFT JOIN public.sat_nums AS s ON s.char_key = chars.char_key
          ), 0)
          AND n.sha_sum = COALESCE((
              SELECT sum(COALESCE(sh.sha_value, 0))
              FROM regexp_split_to_table(n.thname, '') AS chars(char_key)
              LEFT JOIN public.sha_nums AS sh ON sh.char_key = chars.char_key
          ), 0)
  END
  AND NOT EXISTS (
      SELECT 1
      FROM regexp_split_to_table(n.thname, '') AS chars(char_key)
      LEFT JOIN public.sat_nums AS s ON s.char_key = chars.char_key
      LEFT JOIN public.sha_nums AS sh ON sh.char_key = chars.char_key
      WHERE s.char_key IS NULL AND sh.char_key IS NULL
  );

-- dictionary_* is the canonical one-row-per-character calculation. The
-- legacy_ae_expanded_* columns exist only to identify the 4/12 anomaly pattern
-- found in a small legacy subset; they are not an alternative runtime rule.
-- A character present in either dictionary is supported and contributes zero
-- on the missing side. Only a character absent from both is unmapped.
-- The view is intentionally lazy; querying it scans the corpus.
CREATE OR REPLACE VIEW data_governance.v_name_sum_recalculation AS
WITH expanded AS (
    SELECT
        n.name_id,
        n.thname,
        n.sat_sum,
        n.sha_sum,
        split.ch,
        sat.sat_value,
        sha.sha_value
    FROM public.names_miracle AS n
    CROSS JOIN LATERAL regexp_split_to_table(n.thname, '') AS split(ch)
    LEFT JOIN public.sat_nums AS sat ON sat.char_key = split.ch
    LEFT JOIN public.sha_nums AS sha ON sha.char_key = split.ch
)
SELECT
    name_id,
    thname,
    sat_sum AS stored_sat_sum,
    sha_sum AS stored_sha_sum,
    COALESCE(sum(sat_value), 0)::bigint AS dictionary_sat_sum,
    COALESCE(sum(sha_value), 0)::bigint AS dictionary_sha_sum,
    COALESCE(sum(CASE WHEN ch = 'แ' THEN 4 ELSE COALESCE(sat_value, 0) END), 0)::bigint AS legacy_ae_expanded_sat_sum,
    COALESCE(sum(CASE WHEN ch = 'แ' THEN 12 ELSE COALESCE(sha_value, 0) END), 0)::bigint AS legacy_ae_expanded_sha_sum,
    bool_or(ch = 'แ') AS contains_ae,
    count(*) FILTER (WHERE sat_value IS NULL) AS unmapped_sat_character_count,
    count(*) FILTER (WHERE sha_value IS NULL) AS unmapped_sha_character_count,
    count(*) FILTER (WHERE sat_value IS NULL AND sha_value IS NULL) AS unmapped_both_character_count,
    COALESCE(
        array_agg(DISTINCT ch ORDER BY ch) FILTER (WHERE sat_value IS NULL),
        ARRAY[]::text[]
    ) AS unmapped_sat_characters,
    COALESCE(
        array_agg(DISTINCT ch ORDER BY ch) FILTER (WHERE sha_value IS NULL),
        ARRAY[]::text[]
    ) AS unmapped_sha_characters,
    COALESCE(
        array_agg(DISTINCT ch ORDER BY ch) FILTER (WHERE sat_value IS NULL AND sha_value IS NULL),
        ARRAY[]::text[]
    ) AS unmapped_both_characters
FROM expanded
GROUP BY name_id, thname, sat_sum, sha_sum;

CREATE OR REPLACE VIEW data_governance.v_name_sum_mismatches AS
SELECT
    *,
    stored_sat_sum - dictionary_sat_sum AS dictionary_sat_delta,
    stored_sha_sum - dictionary_sha_sum AS dictionary_sha_delta,
    stored_sat_sum - legacy_ae_expanded_sat_sum AS legacy_ae_expanded_sat_delta,
    stored_sha_sum - legacy_ae_expanded_sha_sum AS legacy_ae_expanded_sha_delta,
    (
        contains_ae
        AND stored_sat_sum = legacy_ae_expanded_sat_sum
        AND stored_sha_sum = legacy_ae_expanded_sha_sum
    ) AS matches_legacy_ae_expansion
FROM data_governance.v_name_sum_recalculation
WHERE stored_sat_sum IS DISTINCT FROM dictionary_sat_sum
   OR stored_sha_sum IS DISTINCT FROM dictionary_sha_sum;

-- Prefixes are review signals, not deletion rules. Thai names can legitimately
-- begin with the same characters, so delimiter-free matches remain review_only.
CREATE OR REPLACE VIEW data_governance.v_corpus_title_prefix_candidates AS
WITH prefixes(prefix, precedence) AS (
    VALUES
        ('นางสาว', 10),
        ('เด็กหญิง', 20),
        ('เด็กชาย', 30),
        ('หม่อมหลวง', 40),
        ('หม่อมราชวงศ์', 50),
        ('ศาสตราจารย์', 60),
        ('รองศาสตราจารย์', 70),
        ('ผู้ช่วยศาสตราจารย์', 80),
        ('ดร.', 90),
        ('ดร', 100),
        ('นาง', 110),
        ('นาย', 120),
        ('คุณ', 130),
        ('พระ', 140),
        ('พล.ต.', 150),
        ('พ.ต.', 160),
        ('ร.ต.', 170)
), matches AS (
    SELECT
        n.name_id,
        n.thname,
        p.prefix,
        p.precedence,
        substring(n.thname FROM char_length(p.prefix) + 1 FOR 1) AS next_character
    FROM public.names_miracle AS n
    JOIN prefixes AS p ON n.thname LIKE p.prefix || '%'
)
SELECT DISTINCT ON (name_id)
    name_id,
    thname,
    prefix AS matched_prefix,
    CASE
        WHEN thname = prefix THEN 'high'
        WHEN prefix IN ('เด็กหญิง', 'เด็กชาย') THEN 'high'
        WHEN next_character ~ '^[[:space:].,:;/_-]$' THEN 'high'
        ELSE 'review_only'
    END AS audit_confidence
FROM matches
ORDER BY name_id, char_length(prefix) DESC, precedence;

CREATE OR REPLACE VIEW data_governance.v_phonetic_calibration_stats AS
WITH raw AS (
    SELECT
        COALESCE(NULLIF(btrim(phonetic_version), ''), '<missing>') AS phonetic_version,
        euphony_score,
        pronunciation_ease,
        rhythm_score,
        (
            COALESCE(euphony_score, 0)::numeric * 0.45::numeric
            + COALESCE(pronunciation_ease, 0)::numeric * 0.35::numeric
            + COALESCE(rhythm_score, 0)::numeric * 0.20::numeric
        ) AS raw_sound_score
    FROM data_governance.v_runtime_eligible_name_rows
)
SELECT
    'sound_45_35_20_v1'::text AS formula_version,
    phonetic_version,
    count(*) AS sample_count,
    avg(raw_sound_score) AS sound_mean,
    stddev_samp(raw_sound_score) AS sound_stddev,
    min(raw_sound_score) AS sound_min,
    max(raw_sound_score) AS sound_max,
    percentile_cont(0.10) WITHIN GROUP (ORDER BY raw_sound_score) AS sound_p10,
    percentile_cont(0.50) WITHIN GROUP (ORDER BY raw_sound_score) AS sound_median,
    percentile_cont(0.90) WITHIN GROUP (ORDER BY raw_sound_score) AS sound_p90,
    avg(COALESCE(euphony_score, 0)::numeric) AS euphony_mean,
    avg(COALESCE(pronunciation_ease, 0)::numeric) AS pronunciation_mean,
    avg(COALESCE(rhythm_score, 0)::numeric) AS rhythm_mean,
    count(*) FILTER (
        WHERE euphony_score IS NULL OR euphony_score NOT BETWEEN 0 AND 100
           OR pronunciation_ease IS NULL OR pronunciation_ease NOT BETWEEN 0 AND 100
           OR rhythm_score IS NULL OR rhythm_score NOT BETWEEN 0 AND 100
    ) AS out_of_range_rows
FROM raw
GROUP BY phonetic_version;

-- A read-only calibration surface. It calculates z-scores from current live
-- version statistics; production ranking should use an approved snapshot from
-- phonetic_calibration_snapshots so scores do not drift silently.
CREATE OR REPLACE VIEW data_governance.v_phonetic_live_zscores AS
WITH raw AS (
    SELECT
        n.name_id,
        n.thname,
        COALESCE(NULLIF(btrim(n.phonetic_version), ''), '<missing>') AS phonetic_version,
        (
            COALESCE(n.euphony_score, 0)::numeric * 0.45::numeric
            + COALESCE(n.pronunciation_ease, 0)::numeric * 0.35::numeric
            + COALESCE(n.rhythm_score, 0)::numeric * 0.20::numeric
        ) AS raw_sound_score
    FROM data_governance.v_runtime_eligible_name_rows AS n
)
SELECT
    raw.name_id,
    raw.thname,
    raw.phonetic_version,
    stats.formula_version,
    raw.raw_sound_score,
    stats.sound_mean AS version_sound_mean,
    stats.sound_stddev AS version_sound_stddev,
    (raw.raw_sound_score - stats.sound_mean) / NULLIF(stats.sound_stddev, 0) AS live_sound_zscore
FROM raw
JOIN data_governance.v_phonetic_calibration_stats AS stats
  ON stats.phonetic_version = raw.phonetic_version;

COMMENT ON SCHEMA data_governance IS
    'Version registry and read-only audit/calibration surfaces. Not a replacement for reviewed source data.';

COMMENT ON VIEW data_governance.v_name_sum_recalculation IS
    'Recomputes canonical sums using thai_sum_v1_dictionary and exposes legacy แ=4/12 values only as an anomaly diagnostic.';

COMMENT ON VIEW data_governance.v_character_dictionary_coverage IS
    'Documents one-sided mappings explicitly: a present character is supported and the missing numerology dimension contributes zero.';

COMMENT ON VIEW data_governance.v_runtime_eligible_name_rows IS
    'Day-independent title_and_dictionary_coverage_v2 corpus used for runtime calibration/provenance; request-specific pair and Kalakini gates are applied later.';

COMMENT ON VIEW data_governance.v_phonetic_calibration_stats IS
    'Per-source-version statistics for sound_45_35_20_v1, calculated from live data without hard-coded means/stddevs.';

COMMENT ON TABLE data_governance.semantic_vector_snapshots IS
    'Approval metadata for a versioned candidate-vector corpus. Absence of a matching approved row makes semantic ranking unavailable.';
