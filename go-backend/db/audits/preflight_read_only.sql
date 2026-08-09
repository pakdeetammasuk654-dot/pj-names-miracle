-- Self-contained, read-only preflight for the current production schema.
-- It does not require the data_governance migrations to be installed.

\set ON_ERROR_STOP on
\pset pager off
\pset null '<NULL>'

BEGIN TRANSACTION READ ONLY;
SET LOCAL statement_timeout = '10min';
SET LOCAL lock_timeout = '5s';

\echo '=== Schema and corpus inventory ==='
SELECT version, dirty FROM public.schema_migrations ORDER BY version;

SELECT 'names_miracle' AS table_name, count(*) AS row_count FROM public.names_miracle
UNION ALL SELECT 'numbers', count(*) FROM public.numbers
UNION ALL SELECT 'sat_nums', count(*) FROM public.sat_nums
UNION ALL SELECT 'sha_nums', count(*) FROM public.sha_nums
ORDER BY table_name;

\echo '=== Exact pair-key duplicates (distinct 6 and 06 doctrine) ==='
WITH canonical AS (
    SELECT
        pairnumberid,
        pairnumber AS raw_pairnumber,
        btrim(pairnumber) AS trimmed_pairnumber,
        CASE
            WHEN btrim(pairnumber) ~ '^[0-9]{1,2}$'
            THEN btrim(pairnumber)
        END AS canonical_pair,
        btrim(pairtype) AS pairtype,
        pairpoint
    FROM public.numbers
)
SELECT
    canonical_pair,
    count(*) AS row_count,
    string_agg(
        quote_literal(trimmed_pairnumber) || ':' || pairtype || ':' || pairpoint,
        ', ' ORDER BY char_length(trimmed_pairnumber), trimmed_pairnumber
    ) AS duplicate_rows
FROM canonical
WHERE canonical_pair IS NOT NULL
GROUP BY canonical_pair
HAVING count(*) > 1
ORDER BY canonical_pair;

\echo '=== Explicit distinct 6 and 06 entries ==='
SELECT
    pairnumberid,
    quote_literal(btrim(pairnumber)) AS pair_key,
    btrim(pairtype) AS pairtype,
    pairpoint,
    btrim(miracledesc) AS description
FROM public.numbers
WHERE btrim(pairnumber) IN ('6', '06')
ORDER BY char_length(btrim(pairnumber)), btrim(pairnumber);

\echo '=== Pair grade/point anomalies, including pair 40 ==='
WITH canonical AS (
    SELECT
        pairnumberid,
        pairnumber AS raw_pairnumber,
        CASE
            WHEN btrim(pairnumber) ~ '^[0-9]{1,2}$'
            THEN btrim(pairnumber)
        END AS canonical_pair,
        btrim(pairtype) AS pairtype,
        pairpoint,
        btrim(miracledesc) AS description
    FROM public.numbers
)
SELECT
    pairnumberid,
    quote_literal(btrim(raw_pairnumber)) AS pair_key,
    canonical_pair,
    pairtype,
    pairpoint,
    description,
    concat_ws(
        ', ',
        CASE WHEN pairtype LIKE 'D%' AND pairpoint = 0 THEN 'zero_d_point' END,
        CASE WHEN pairtype LIKE 'R%' AND pairpoint >= 0 THEN 'non_negative_r_point' END,
        CASE WHEN canonical_pair IS NULL THEN 'invalid_pair_key' END
    ) AS anomaly_codes
FROM canonical
WHERE (pairtype LIKE 'D%' AND pairpoint = 0)
   OR (pairtype LIKE 'R%' AND pairpoint >= 0)
   OR canonical_pair IS NULL
ORDER BY canonical_pair, pairnumberid;

\echo '=== Character dictionary coverage (one-sided mappings are supported) ==='
WITH coverage AS (
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
    FULL OUTER JOIN public.sha_nums AS sha ON sha.char_key = sat.char_key
)
SELECT
    mapping_scope,
    count(*) AS character_count,
    string_agg(char_key, '' ORDER BY char_key) AS characters
FROM coverage
GROUP BY mapping_scope
ORDER BY mapping_scope;

\echo '=== Stored sums versus canonical character dictionaries ==='
WITH expanded AS MATERIALIZED (
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
), recalculated AS (
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
        count(*) FILTER (WHERE sat_value IS NULL) AS unmapped_sat_count,
        count(*) FILTER (WHERE sha_value IS NULL) AS unmapped_sha_count,
        count(*) FILTER (WHERE sat_value IS NULL AND sha_value IS NULL) AS unmapped_both_count
    FROM expanded
    GROUP BY name_id, thname, sat_sum, sha_sum
)
SELECT
    count(*) AS names,
    count(*) FILTER (
        WHERE stored_sat_sum IS DISTINCT FROM dictionary_sat_sum
           OR stored_sha_sum IS DISTINCT FROM dictionary_sha_sum
    ) AS dictionary_mismatch_names,
    count(*) FILTER (
        WHERE contains_ae
          AND stored_sat_sum = legacy_ae_expanded_sat_sum
          AND stored_sha_sum = legacy_ae_expanded_sha_sum
          AND (
              stored_sat_sum IS DISTINCT FROM dictionary_sat_sum
              OR stored_sha_sum IS DISTINCT FROM dictionary_sha_sum
          )
    ) AS legacy_ae_expansion_anomaly_names,
    count(*) FILTER (WHERE contains_ae) AS names_containing_ae,
    count(*) FILTER (WHERE unmapped_sat_count > 0) AS names_with_unmapped_sat,
    count(*) FILTER (WHERE unmapped_sha_count > 0) AS names_with_unmapped_sha,
    count(*) FILTER (WHERE unmapped_both_count > 0) AS names_with_unmapped_both
FROM recalculated;

\echo '=== Canonical sum mismatch samples; legacy แ=4/12 pattern is diagnostic only ==='
WITH expanded AS MATERIALIZED (
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
), recalculated AS (
    SELECT
        name_id,
        thname,
        sat_sum AS stored_sat_sum,
        sha_sum AS stored_sha_sum,
        COALESCE(sum(sat_value), 0)::bigint AS dictionary_sat_sum,
        COALESCE(sum(sha_value), 0)::bigint AS dictionary_sha_sum,
        COALESCE(sum(CASE WHEN ch = 'แ' THEN 4 ELSE COALESCE(sat_value, 0) END), 0)::bigint AS legacy_ae_expanded_sat_sum,
        COALESCE(sum(CASE WHEN ch = 'แ' THEN 12 ELSE COALESCE(sha_value, 0) END), 0)::bigint AS legacy_ae_expanded_sha_sum,
        bool_or(ch = 'แ') AS contains_ae
    FROM expanded
    GROUP BY name_id, thname, sat_sum, sha_sum
)
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
FROM recalculated
WHERE stored_sat_sum IS DISTINCT FROM dictionary_sat_sum
   OR stored_sha_sum IS DISTINCT FROM dictionary_sha_sum
ORDER BY contains_ae DESC, name_id
LIMIT 100;

\echo '=== Corpus title-prefix audit (review/quarantine candidates only) ==='
WITH prefixes(prefix, precedence) AS (
    VALUES
        ('นางสาว', 10), ('เด็กหญิง', 20), ('เด็กชาย', 30),
        ('หม่อมหลวง', 40), ('หม่อมราชวงศ์', 50),
        ('ศาสตราจารย์', 60), ('รองศาสตราจารย์', 70),
        ('ผู้ช่วยศาสตราจารย์', 80), ('ดร.', 90), ('ดร', 100),
        ('นาง', 110), ('นาย', 120), ('คุณ', 130), ('พระ', 140),
        ('พล.ต.', 150), ('พ.ต.', 160), ('ร.ต.', 170)
), matches AS (
    SELECT
        n.name_id,
        n.thname,
        p.prefix,
        p.precedence,
        substring(n.thname FROM char_length(p.prefix) + 1 FOR 1) AS next_character
    FROM public.names_miracle AS n
    JOIN prefixes AS p ON n.thname LIKE p.prefix || '%'
), candidates AS (
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
    ORDER BY name_id, char_length(prefix) DESC, precedence
)
SELECT matched_prefix, audit_confidence, count(*) AS candidate_count
FROM candidates
GROUP BY matched_prefix, audit_confidence
ORDER BY candidate_count DESC, matched_prefix;

\echo '=== First 100 title-prefix candidates (never auto-delete from this report) ==='
WITH prefixes(prefix, precedence) AS (
    VALUES
        ('นางสาว', 10), ('เด็กหญิง', 20), ('เด็กชาย', 30),
        ('หม่อมหลวง', 40), ('หม่อมราชวงศ์', 50),
        ('ศาสตราจารย์', 60), ('รองศาสตราจารย์', 70),
        ('ผู้ช่วยศาสตราจารย์', 80), ('ดร.', 90), ('ดร', 100),
        ('นาง', 110), ('นาย', 120), ('คุณ', 130), ('พระ', 140),
        ('พล.ต.', 150), ('พ.ต.', 160), ('ร.ต.', 170)
), matches AS (
    SELECT
        n.name_id,
        n.thname,
        p.prefix,
        p.precedence,
        substring(n.thname FROM char_length(p.prefix) + 1 FOR 1) AS next_character
    FROM public.names_miracle AS n
    JOIN prefixes AS p ON n.thname LIKE p.prefix || '%'
), candidates AS (
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
    ORDER BY name_id, char_length(prefix) DESC, precedence
)
SELECT *
FROM candidates
ORDER BY audit_confidence, matched_prefix, name_id
LIMIT 100;

\echo '=== Exact sound formula calibration by phonetic source version ==='
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
    FROM public.names_miracle
)
SELECT
    'sound_45_35_20_v1' AS formula_version,
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
GROUP BY phonetic_version
ORDER BY phonetic_version;

\echo '=== Existing governance constraints and validation state ==='
SELECT
    c.conrelid::regclass AS table_name,
    c.conname AS constraint_name,
    c.contype AS constraint_type,
    c.convalidated AS is_validated
FROM pg_constraint AS c
WHERE c.conrelid IN (
    'public.names_miracle'::regclass,
    'public.numbers'::regclass,
    'public.sat_nums'::regclass,
    'public.sha_nums'::regclass
)
ORDER BY c.conrelid::regclass::text, c.conname;

ROLLBACK;
