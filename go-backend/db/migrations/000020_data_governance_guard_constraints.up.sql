-- Guard future writes without rewriting or deleting existing rows.
-- NOT VALID means PostgreSQL enforces each CHECK for new/updated rows while
-- existing violations remain available for audit and controlled remediation.

DO $migration$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'numbers_pairnumber_two_digit_ck'
          AND conrelid = 'public.numbers'::regclass
    ) THEN
        ALTER TABLE public.numbers
            ADD CONSTRAINT numbers_pairnumber_two_digit_ck
            CHECK (
                pairnumber IS NOT NULL
                AND btrim(pairnumber) ~ '^[0-9]{1,2}$'
            ) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'numbers_pairtype_known_ck'
          AND conrelid = 'public.numbers'::regclass
    ) THEN
        ALTER TABLE public.numbers
            ADD CONSTRAINT numbers_pairtype_known_ck
            CHECK (
                pairtype IS NOT NULL
                AND btrim(pairtype) IN ('D10', 'D8', 'D5', 'R10', 'R7', 'R5')
            ) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'numbers_pairpoint_bounded_ck'
          AND conrelid = 'public.numbers'::regclass
    ) THEN
        ALTER TABLE public.numbers
            ADD CONSTRAINT numbers_pairpoint_bounded_ck
            CHECK (pairpoint IS NOT NULL AND pairpoint BETWEEN -100 AND 100) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'numbers_grade_point_sign_ck'
          AND conrelid = 'public.numbers'::regclass
    ) THEN
        ALTER TABLE public.numbers
            ADD CONSTRAINT numbers_grade_point_sign_ck
            CHECK (
                (btrim(pairtype) LIKE 'D%' AND pairpoint <> 0)
                OR (btrim(pairtype) LIKE 'R%' AND pairpoint < 0)
            ) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'names_miracle_sums_present_ck'
          AND conrelid = 'public.names_miracle'::regclass
    ) THEN
        ALTER TABLE public.names_miracle
            ADD CONSTRAINT names_miracle_sums_present_ck
            CHECK (
                sat_sum IS NOT NULL AND sat_sum >= 0
                AND sha_sum IS NOT NULL AND sha_sum >= 0
            ) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'names_miracle_sound_components_range_ck'
          AND conrelid = 'public.names_miracle'::regclass
    ) THEN
        ALTER TABLE public.names_miracle
            ADD CONSTRAINT names_miracle_sound_components_range_ck
            CHECK (
                euphony_score IS NOT NULL AND euphony_score BETWEEN 0 AND 100
                AND pronunciation_ease IS NOT NULL AND pronunciation_ease BETWEEN 0 AND 100
                AND rhythm_score IS NOT NULL AND rhythm_score BETWEEN 0 AND 100
            ) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'names_miracle_phonetic_version_present_ck'
          AND conrelid = 'public.names_miracle'::regclass
    ) THEN
        ALTER TABLE public.names_miracle
            ADD CONSTRAINT names_miracle_phonetic_version_present_ck
            CHECK (phonetic_version IS NOT NULL AND btrim(phonetic_version) <> '') NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'sat_nums_single_character_ck'
          AND conrelid = 'public.sat_nums'::regclass
    ) THEN
        ALTER TABLE public.sat_nums
            ADD CONSTRAINT sat_nums_single_character_ck
            CHECK (char_length(char_key) = 1 AND sat_value >= 0) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'sha_nums_single_character_ck'
          AND conrelid = 'public.sha_nums'::regclass
    ) THEN
        ALTER TABLE public.sha_nums
            ADD CONSTRAINT sha_nums_single_character_ck
            CHECK (char_length(char_key) = 1 AND sha_value >= 0) NOT VALID;
    END IF;
END
$migration$;

COMMENT ON CONSTRAINT numbers_pairnumber_two_digit_ck ON public.numbers IS
    'New/updated pair keys must use authored 0-99 digit text. Single-digit keys and leading-zero keys are distinct.';

COMMENT ON CONSTRAINT numbers_grade_point_sign_ck ON public.numbers IS
    'New/updated D rows require non-zero points and R rows require negative points; zero is not a valid classified score.';

COMMENT ON CONSTRAINT names_miracle_sound_components_range_ck ON public.names_miracle IS
    'Raw sound components must remain on the declared 0-100 scale; source-version calibration is audited separately.';
