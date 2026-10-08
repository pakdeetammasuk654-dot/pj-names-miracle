-- Read-only completeness audit for public.names_miracle.
\set ON_ERROR_STOP on
\pset pager off

BEGIN READ ONLY;
SET LOCAL statement_timeout = '10min';

SELECT
    count(*) AS total_rows,
    count(*) FILTER (WHERE thname IS NULL OR btrim(thname) = '') AS missing_name,
    count(*) FILTER (WHERE meaning IS NULL OR btrim(meaning) = '') AS missing_meaning,
    count(*) FILTER (WHERE meaning_vector IS NULL) AS missing_vector,
    count(*) FILTER (WHERE meaning_vector IS NOT NULL AND vector_dims(meaning_vector) <> 1536) AS wrong_vector_dimensions,
    count(*) FILTER (WHERE sat_sum IS NULL OR sha_sum IS NULL) AS missing_sums,
    count(*) FILTER (WHERE satnum IS NULL OR cardinality(satnum) = 0) AS missing_satnum,
    count(*) FILTER (WHERE shanum IS NULL OR cardinality(shanum) = 0) AS missing_shanum,
    count(*) FILTER (WHERE t_sat IS NULL OR cardinality(t_sat) = 0) AS missing_t_sat,
    count(*) FILTER (WHERE t_sha IS NULL OR cardinality(t_sha) = 0) AS missing_t_sha,
    count(*) FILTER (WHERE gender IS NULL OR btrim(gender) = '') AS missing_gender,
    count(*) FILTER (WHERE gender IS NOT NULL AND btrim(gender) NOT IN ('ช', 'ญ', 'ค')) AS invalid_gender,
    count(*) FILTER (WHERE NOT phonetic_processed) AS phonetic_not_processed,
    count(*) FILTER (
        WHERE pronunciation_ease IS NULL OR euphony_score IS NULL OR rhythm_score IS NULL
    ) AS missing_phonetic_scores
FROM public.names_miracle;

SELECT name_id, thname, meaning, gender
FROM public.names_miracle
WHERE meaning IS NULL OR btrim(meaning) = ''
   OR meaning_vector IS NULL
   OR gender IS NULL OR btrim(gender) = ''
   OR btrim(gender) NOT IN ('ช', 'ญ', 'ค')
ORDER BY name_id;

COMMIT;
