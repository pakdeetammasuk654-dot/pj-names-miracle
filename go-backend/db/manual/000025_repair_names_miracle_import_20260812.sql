-- Deterministic repair for the names_ubu_2026_08 import and known canonical
-- sum/gender anomalies. Run with `psql -v apply=0 -f ...` for a rollback-only
-- rehearsal, then `-v apply=1` after reviewing all assertions and counts.
\set ON_ERROR_STOP on
\if :{?apply}
\else
  \set apply 0
\endif

BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

CREATE SCHEMA IF NOT EXISTS data_governance;

CREATE TABLE IF NOT EXISTS data_governance.names_miracle_repair_20260812_backup
(LIKE public.names_miracle INCLUDING DEFAULTS INCLUDING CONSTRAINTS);

CREATE TEMP TABLE canonical_ae_repairs ON COMMIT DROP AS
WITH recalculated AS (
    SELECT n.name_id,
        coalesce(sum(s.sat_value), 0)::integer AS sat_sum,
        coalesce(sum(sh.sha_value), 0)::integer AS sha_sum
    FROM public.names_miracle AS n
    CROSS JOIN LATERAL regexp_split_to_table(n.thname, '') chars(ch)
    LEFT JOIN public.sat_nums AS s ON s.char_key = chars.ch
    LEFT JOIN public.sha_nums AS sh ON sh.char_key = chars.ch
    WHERE strpos(n.thname, 'แ') > 0
    GROUP BY n.name_id
)
SELECT r.*
FROM recalculated AS r
JOIN public.names_miracle AS n USING (name_id)
WHERE n.sat_sum IS DISTINCT FROM r.sat_sum
   OR n.sha_sum IS DISTINCT FROM r.sha_sum;

DO $check$
DECLARE
    missing_import_arrays bigint;
BEGIN
    SELECT count(*) INTO missing_import_arrays
    FROM data_governance.name_import_staging AS s
    JOIN data_governance.name_import_batches AS b USING (batch_id)
    JOIN public.names_miracle AS n ON n.name_id = s.promoted_name_id
    WHERE b.batch_label = 'names_ubu_2026_08'
      AND (
          n.satnum IS NULL OR cardinality(n.satnum) = 0
          OR n.shanum IS NULL OR cardinality(n.shanum) = 0
          OR n.t_sat IS NULL OR cardinality(n.t_sat) = 0
          OR n.t_sha IS NULL OR cardinality(n.t_sha) = 0
      );
    IF missing_import_arrays <> 490 THEN
        RAISE EXCEPTION 'expected 490 imported rows with missing arrays, found %', missing_import_arrays;
    END IF;
END
$check$;

INSERT INTO data_governance.names_miracle_repair_20260812_backup
SELECT DISTINCT n.*
FROM public.names_miracle AS n
LEFT JOIN data_governance.name_import_staging AS s ON s.promoted_name_id = n.name_id
LEFT JOIN data_governance.name_import_batches AS b ON b.batch_id = s.batch_id
WHERE (
        b.batch_label = 'names_ubu_2026_08'
        AND (
            n.satnum IS NULL OR cardinality(n.satnum) = 0
            OR n.shanum IS NULL OR cardinality(n.shanum) = 0
            OR n.t_sat IS NULL OR cardinality(n.t_sat) = 0
            OR n.t_sha IS NULL OR cardinality(n.t_sha) = 0
        )
      )
   OR n.name_id = 461658
   OR btrim(n.gender) IN ('ชาย', 'หญิง', 'ทุกเพศ - ค', 'ค -', '#ช', '##ช')
   OR EXISTS (SELECT 1 FROM canonical_ae_repairs AS r WHERE r.name_id = n.name_id)
ON CONFLICT DO NOTHING;

WITH target AS (
    SELECT n.name_id, n.sat_sum, n.sha_sum
    FROM data_governance.name_import_staging AS s
    JOIN data_governance.name_import_batches AS b USING (batch_id)
    JOIN public.names_miracle AS n ON n.name_id = s.promoted_name_id
    WHERE b.batch_label = 'names_ubu_2026_08'
      AND (
          n.satnum IS NULL OR cardinality(n.satnum) = 0
          OR n.shanum IS NULL OR cardinality(n.shanum) = 0
          OR n.t_sat IS NULL OR cardinality(n.t_sat) = 0
          OR n.t_sha IS NULL OR cardinality(n.t_sha) = 0
      )
), expanded AS (
    SELECT t.name_id, kind, pos,
        CASE
            WHEN length(total::text) = 1 THEN lpad(total::text, 2, '0')
            ELSE substr(total::text, pos, 2)
        END AS pair
    FROM target AS t
    CROSS JOIN LATERAL (VALUES ('sat', t.sat_sum), ('sha', t.sha_sum)) sums(kind, total)
    CROSS JOIN LATERAL generate_series(
        1,
        CASE WHEN length(total::text) <= 2 THEN 1 ELSE length(total::text) - 1 END
    ) positions(pos)
    WHERE total > 0
), rebuilt AS (
    SELECT name_id,
        array_agg(pair ORDER BY pos) FILTER (WHERE kind = 'sat') AS sat_pairs,
        array_agg(pair ORDER BY pos) FILTER (WHERE kind = 'sha') AS sha_pairs
    FROM expanded
    GROUP BY name_id
)
UPDATE public.names_miracle AS n
SET satnum = r.sat_pairs,
    shanum = r.sha_pairs,
    t_sat = r.sat_pairs,
    t_sha = r.sha_pairs
FROM rebuilt AS r
WHERE n.name_id = r.name_id;

UPDATE public.names_miracle
SET t_sat = satnum::text[], t_sha = shanum::text[]
WHERE name_id = 461658
  AND satnum IS NOT NULL AND shanum IS NOT NULL
  AND (t_sat IS NULL OR t_sha IS NULL);

UPDATE public.names_miracle
SET gender = CASE btrim(gender)
    WHEN 'ชาย' THEN 'ช'
    WHEN 'หญิง' THEN 'ญ'
    WHEN 'ทุกเพศ - ค' THEN 'ค'
    WHEN 'ค -' THEN 'ค'
    WHEN '#ช' THEN 'ช'
    WHEN '##ช' THEN 'ช'
    ELSE gender
END
WHERE btrim(gender) IN ('ชาย', 'หญิง', 'ทุกเพศ - ค', 'ค -', '#ช', '##ช');

-- Correct the known historical doubled-แ totals from the canonical character
-- dictionaries, then rebuild their pair arrays from the corrected totals.
UPDATE public.names_miracle AS n
SET sat_sum = r.sat_sum, sha_sum = r.sha_sum
FROM canonical_ae_repairs AS r
WHERE n.name_id = r.name_id
  AND (n.sat_sum IS DISTINCT FROM r.sat_sum OR n.sha_sum IS DISTINCT FROM r.sha_sum);

WITH expanded AS (
    SELECT c.name_id, kind, pos,
        CASE WHEN length(total::text) = 1 THEN lpad(total::text, 2, '0') ELSE substr(total::text, pos, 2) END pair
    FROM canonical_ae_repairs AS c
    CROSS JOIN LATERAL (VALUES ('sat', c.sat_sum), ('sha', c.sha_sum)) sums(kind, total)
    CROSS JOIN LATERAL generate_series(1, CASE WHEN length(total::text) <= 2 THEN 1 ELSE length(total::text)-1 END) positions(pos)
), rebuilt AS (
    SELECT name_id,
        array_agg(pair ORDER BY pos) FILTER (WHERE kind='sat') sat_pairs,
        array_agg(pair ORDER BY pos) FILTER (WHERE kind='sha') sha_pairs
    FROM expanded GROUP BY name_id
)
UPDATE public.names_miracle AS n
SET satnum=r.sat_pairs, shanum=r.sha_pairs, t_sat=r.sat_pairs, t_sha=r.sha_pairs
FROM rebuilt AS r WHERE n.name_id=r.name_id;

DO $verify$
DECLARE
    remaining_arrays bigint;
    remaining_normalizable_gender bigint;
    remaining_ae_sum_mismatches bigint;
BEGIN
    SELECT count(*) INTO remaining_arrays
    FROM data_governance.name_import_staging AS s
    JOIN data_governance.name_import_batches AS b USING (batch_id)
    JOIN public.names_miracle AS n ON n.name_id = s.promoted_name_id
    WHERE b.batch_label = 'names_ubu_2026_08'
      AND (n.satnum IS NULL OR n.shanum IS NULL OR n.t_sat IS NULL OR n.t_sha IS NULL);
    IF remaining_arrays <> 0 THEN
        RAISE EXCEPTION 'repair incomplete: % imported rows still lack arrays', remaining_arrays;
    END IF;

    SELECT count(*) INTO remaining_normalizable_gender
    FROM public.names_miracle
    WHERE btrim(gender) IN ('ชาย', 'หญิง', 'ทุกเพศ - ค', 'ค -', '#ช', '##ช');
    IF remaining_normalizable_gender <> 0 THEN
        RAISE EXCEPTION 'repair incomplete: % deterministic gender values remain', remaining_normalizable_gender;
    END IF;

    SELECT count(*) INTO remaining_ae_sum_mismatches
    FROM canonical_ae_repairs AS r
    JOIN public.names_miracle AS n USING (name_id)
    WHERE n.sat_sum IS DISTINCT FROM r.sat_sum
       OR n.sha_sum IS DISTINCT FROM r.sha_sum;
    IF remaining_ae_sum_mismatches <> 0 THEN
        RAISE EXCEPTION 'repair incomplete: % canonical แ sum mismatches remain', remaining_ae_sum_mismatches;
    END IF;
END
$verify$;

\if :apply
  COMMIT;
  \echo 'Repair committed.'
\else
  ROLLBACK;
  \echo 'Dry run rolled back. Re-run with -v apply=1 to commit.'
\endif
