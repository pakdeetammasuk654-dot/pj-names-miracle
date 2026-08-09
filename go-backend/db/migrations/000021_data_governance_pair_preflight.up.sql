-- A duplicate-tolerant lookup index plus a reusable assertion for the later
-- exact key uniqueness rollout. Single-digit keys and leading-zero keys are
-- intentionally distinct domain entries.

CREATE INDEX IF NOT EXISTS numbers_canonical_pair_lookup_idx
ON public.numbers (
    (btrim(pairnumber))
)
WHERE btrim(pairnumber) ~ '^[0-9]{1,2}$';

CREATE OR REPLACE FUNCTION data_governance.assert_canonical_pairs_ready()
RETURNS void
LANGUAGE plpgsql
AS $function$
DECLARE
    invalid_count bigint;
    duplicate_count bigint;
BEGIN
    SELECT count(*)
    INTO duplicate_count
    FROM (
        SELECT btrim(pairnumber) AS pair_key
        FROM public.numbers
        WHERE btrim(pairnumber) ~ '^[0-9]{1,2}$'
        GROUP BY btrim(pairnumber)
        HAVING count(*) > 1
    ) AS duplicates;

    IF duplicate_count > 0 THEN
        RAISE EXCEPTION
            'pair key enforcement blocked: % exact pair keys are duplicated',
            duplicate_count;
    END IF;

    SELECT count(*)
    INTO invalid_count
    FROM public.numbers
    WHERE pairnumber IS NULL
       OR btrim(pairnumber) !~ '^[0-9]{1,2}$';

    IF invalid_count > 0 THEN
        RAISE EXCEPTION
            'pair key enforcement blocked: % pair rows are not one- or two-digit keys',
            invalid_count;
    END IF;
END
$function$;

COMMENT ON FUNCTION data_governance.assert_canonical_pairs_ready() IS
    'Read-only preflight. Raises before exact pair-key uniqueness is enforced if malformed keys or exact duplicates remain.';
