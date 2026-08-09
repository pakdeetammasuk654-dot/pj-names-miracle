DROP TRIGGER IF EXISTS names_miracle_vector_snapshot_link_trg ON public.names_miracle;
DROP TRIGGER IF EXISTS semantic_vector_snapshot_approval_trg ON data_governance.semantic_vector_snapshots;
DROP TRIGGER IF EXISTS semantic_vector_snapshot_immutable_trg ON data_governance.semantic_vector_snapshots;
DROP TRIGGER IF EXISTS phonetic_calibration_snapshot_immutable_trg ON data_governance.phonetic_calibration_snapshots;
DROP FUNCTION IF EXISTS data_governance.validate_semantic_snapshot_approval();
DROP FUNCTION IF EXISTS data_governance.guard_meaning_vector_snapshot_link();
DROP FUNCTION IF EXISTS data_governance.reject_approved_snapshot_mutation();
ALTER TABLE public.names_miracle
    DROP CONSTRAINT IF EXISTS names_miracle_meaning_vector_snapshot_fk;
ALTER TABLE public.names_miracle
    DROP COLUMN IF EXISTS meaning_vector_snapshot_id;
