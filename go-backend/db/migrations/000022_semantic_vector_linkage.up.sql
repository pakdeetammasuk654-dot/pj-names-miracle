-- Link every usable candidate vector to one governed snapshot. Existing rows
-- remain NULL/unapproved; adding this nullable column does not rewrite or bless
-- the legacy vector corpus.

ALTER TABLE public.names_miracle
    ADD COLUMN IF NOT EXISTS meaning_vector_snapshot_id bigint;

DO $migration$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'names_miracle_meaning_vector_snapshot_fk'
          AND conrelid = 'public.names_miracle'::regclass
    ) THEN
        ALTER TABLE public.names_miracle
            ADD CONSTRAINT names_miracle_meaning_vector_snapshot_fk
            FOREIGN KEY (meaning_vector_snapshot_id)
            REFERENCES data_governance.semantic_vector_snapshots(snapshot_id)
            NOT VALID;
    END IF;
END
$migration$;

-- Approved governance evidence is append-only. A correction is a new snapshot,
-- never an in-place rewrite of the record that justified a delivered rank.
CREATE OR REPLACE FUNCTION data_governance.reject_approved_snapshot_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    IF OLD.approved_at IS NOT NULL THEN
        RAISE EXCEPTION 'approved snapshot rows are immutable; create a new snapshot';
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS semantic_vector_snapshot_immutable_trg
ON data_governance.semantic_vector_snapshots;
CREATE TRIGGER semantic_vector_snapshot_immutable_trg
BEFORE UPDATE OR DELETE ON data_governance.semantic_vector_snapshots
FOR EACH ROW EXECUTE FUNCTION data_governance.reject_approved_snapshot_mutation();

DROP TRIGGER IF EXISTS phonetic_calibration_snapshot_immutable_trg
ON data_governance.phonetic_calibration_snapshots;
CREATE TRIGGER phonetic_calibration_snapshot_immutable_trg
BEFORE UPDATE OR DELETE ON data_governance.phonetic_calibration_snapshots
FOR EACH ROW EXECUTE FUNCTION data_governance.reject_approved_snapshot_mutation();

-- A vector/name/meaning change invalidates its old proof. Rows may be attached
-- only to a pending snapshot; approval happens after regeneration, fingerprint,
-- count/dimension checks, and held-out evaluation.
CREATE OR REPLACE FUNCTION data_governance.guard_meaning_vector_snapshot_link()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
DECLARE
    target_approved_at timestamptz;
BEGIN
    IF NEW.meaning_vector_snapshot_id IS NOT NULL
       AND (
           TG_OP = 'INSERT'
           OR NEW.meaning_vector_snapshot_id IS DISTINCT FROM OLD.meaning_vector_snapshot_id
       ) THEN
        SELECT approved_at
        INTO target_approved_at
        FROM data_governance.semantic_vector_snapshots
        WHERE snapshot_id = NEW.meaning_vector_snapshot_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'unknown semantic vector snapshot id %', NEW.meaning_vector_snapshot_id;
        END IF;
        IF target_approved_at IS NOT NULL THEN
            RAISE EXCEPTION 'cannot newly attach a row to an already-approved semantic snapshot';
        END IF;
    END IF;

    IF TG_OP = 'UPDATE'
       AND OLD.meaning_vector_snapshot_id IS NOT NULL
       AND NEW.meaning_vector_snapshot_id IS NOT DISTINCT FROM OLD.meaning_vector_snapshot_id
       AND (
           NEW.meaning_vector::text IS DISTINCT FROM OLD.meaning_vector::text
           OR NEW.meaning IS DISTINCT FROM OLD.meaning
           OR NEW.thname IS DISTINCT FROM OLD.thname
           OR NEW.sat_sum IS DISTINCT FROM OLD.sat_sum
           OR NEW.sha_sum IS DISTINCT FROM OLD.sha_sum
           OR NEW.k_sunday IS DISTINCT FROM OLD.k_sunday
           OR NEW.k_monday IS DISTINCT FROM OLD.k_monday
           OR NEW.k_tuesday IS DISTINCT FROM OLD.k_tuesday
           OR NEW.k_wednesday1 IS DISTINCT FROM OLD.k_wednesday1
           OR NEW.k_wednesday2 IS DISTINCT FROM OLD.k_wednesday2
           OR NEW.k_thursday IS DISTINCT FROM OLD.k_thursday
           OR NEW.k_friday IS DISTINCT FROM OLD.k_friday
           OR NEW.k_saturday IS DISTINCT FROM OLD.k_saturday
       ) THEN
        RAISE EXCEPTION 'vector/name/meaning changed without switching to a pending semantic snapshot';
    END IF;

    RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS names_miracle_vector_snapshot_link_trg
ON public.names_miracle;
CREATE TRIGGER names_miracle_vector_snapshot_link_trg
BEFORE INSERT OR UPDATE OF
    meaning_vector, meaning, thname, sat_sum, sha_sum,
    k_sunday, k_monday, k_tuesday, k_wednesday1, k_wednesday2,
    k_thursday, k_friday, k_saturday, meaning_vector_snapshot_id
ON public.names_miracle
FOR EACH ROW EXECUTE FUNCTION data_governance.guard_meaning_vector_snapshot_link();

-- Validate live linkage at the moment approval is granted. Runtime repeats the
-- inexpensive count/dimension checks when it loads the snapshot; the recorded
-- SHA-256 and held-out evaluation remain reviewed governance evidence.
CREATE OR REPLACE FUNCTION data_governance.validate_semantic_snapshot_approval()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
DECLARE
    actual_corpus_rows bigint;
    actual_vectorized_rows bigint;
    linked_vector_rows bigint;
    invalid_linked_rows bigint;
BEGIN
    IF TG_OP = 'INSERT' AND NEW.approved_at IS NOT NULL THEN
        RAISE EXCEPTION 'insert semantic snapshots as pending, link/regenerate rows, then approve by update';
    END IF;
    IF TG_OP = 'INSERT' OR NEW.approved_at IS NULL OR OLD.approved_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT
        count(*),
        count(n.meaning_vector),
        count(*) FILTER (
            WHERE n.meaning_vector_snapshot_id = NEW.snapshot_id
              AND n.meaning_vector IS NOT NULL
        ),
        count(*) FILTER (
            WHERE n.meaning_vector_snapshot_id = NEW.snapshot_id
              AND (
                  n.meaning_vector IS NULL
                  OR vector_dims(n.meaning_vector) <> NEW.vector_dimensions
              )
        )
    INTO
        actual_corpus_rows,
        actual_vectorized_rows,
        linked_vector_rows,
        invalid_linked_rows
    FROM public.names_miracle AS n
    JOIN data_governance.v_runtime_eligible_name_rows AS eligible
      ON eligible.name_id = n.name_id;

    IF actual_corpus_rows <> NEW.corpus_row_count
       OR actual_vectorized_rows <> NEW.vectorized_row_count
       OR NEW.vectorized_row_count <> NEW.corpus_row_count
       OR linked_vector_rows <> NEW.vectorized_row_count
       OR invalid_linked_rows <> 0 THEN
        RAISE EXCEPTION
            'semantic snapshot approval counts/dimensions do not match live corpus (corpus %/%, vectors %/%, linked %, invalid %)',
            actual_corpus_rows, NEW.corpus_row_count,
            actual_vectorized_rows, NEW.vectorized_row_count,
            linked_vector_rows, invalid_linked_rows;
    END IF;

    RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS semantic_vector_snapshot_approval_trg
ON data_governance.semantic_vector_snapshots;
CREATE TRIGGER semantic_vector_snapshot_approval_trg
BEFORE INSERT OR UPDATE OF approved_at ON data_governance.semantic_vector_snapshots
FOR EACH ROW EXECUTE FUNCTION data_governance.validate_semantic_snapshot_approval();

COMMENT ON COLUMN public.names_miracle.meaning_vector_snapshot_id IS
    'Pending/approved provenance record for meaning_vector. NULL means semantic candidate-vector provenance is unverified.';
