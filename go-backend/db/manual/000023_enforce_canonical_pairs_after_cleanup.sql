-- MANUAL / POST-REMEDIATION ONLY.
--
-- Do not include this file in an automatic migration run. First resolve every
-- exact duplicate or malformed key reported by the pair-key preflight through
-- an approved data-change migration. Keys such as 6 and 06 are distinct.

\set ON_ERROR_STOP on

BEGIN;

SELECT data_governance.assert_canonical_pairs_ready();

CREATE UNIQUE INDEX IF NOT EXISTS numbers_pair_key_uq
ON public.numbers ((btrim(pairnumber)));

ALTER TABLE public.numbers
    VALIDATE CONSTRAINT numbers_pairnumber_two_digit_ck;
ALTER TABLE public.numbers
    VALIDATE CONSTRAINT numbers_pairtype_known_ck;
ALTER TABLE public.numbers
    VALIDATE CONSTRAINT numbers_pairpoint_bounded_ck;
ALTER TABLE public.numbers
    VALIDATE CONSTRAINT numbers_grade_point_sign_ck;

COMMIT;
