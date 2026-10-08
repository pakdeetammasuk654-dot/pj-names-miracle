-- Stage new names outside the ranking corpus until every required field has
-- been generated and reviewed. This deliberately does not write to
-- public.names_miracle.

CREATE SCHEMA IF NOT EXISTS data_governance;

CREATE TABLE IF NOT EXISTS data_governance.name_import_batches (
    batch_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    batch_label text NOT NULL UNIQUE,
    source_sha256 text NOT NULL CHECK (source_sha256 ~ '^[0-9a-f]{64}$'),
    source_row_count integer NOT NULL CHECK (source_row_count > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by text NOT NULL,
    CHECK (btrim(batch_label) <> ''),
    CHECK (btrim(created_by) <> '')
);

CREATE TABLE IF NOT EXISTS data_governance.name_import_staging (
    staging_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    batch_id bigint NOT NULL REFERENCES data_governance.name_import_batches(batch_id) ON DELETE RESTRICT,
    source_line integer NOT NULL CHECK (source_line > 0),
    thname text NOT NULL,
    normalized_name text NOT NULL,
    stage_status text NOT NULL DEFAULT 'staged'
        CHECK (stage_status IN ('staged', 'meaning_ready', 'phonetic_ready', 'vector_ready', 'review_required', 'promoted', 'rejected')),
    rejection_reason text,
    sat_sum integer,
    sha_sum integer,
    k_sunday boolean,
    k_monday boolean,
    k_tuesday boolean,
    k_wednesday1 boolean,
    k_wednesday2 boolean,
    k_thursday boolean,
    k_friday boolean,
    k_saturday boolean,
    meaning text,
    gender text CHECK (gender IN ('ช', 'ญ', 'ค')),
    meaning_model text,
    meaning_prompt_version text,
    euphony_score integer CHECK (euphony_score BETWEEN 0 AND 100),
    pronunciation_ease integer CHECK (pronunciation_ease BETWEEN 0 AND 100),
    rhythm_score integer CHECK (rhythm_score BETWEEN 0 AND 100),
    phonetic_version text,
    phonetic_model text,
    meaning_vector vector(1536),
    embedding_model text,
    embedding_input_template text,
    reviewed_at timestamptz,
    reviewed_by text,
    promoted_name_id bigint,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (batch_id, normalized_name),
    CHECK (btrim(thname) <> ''),
    CHECK (btrim(normalized_name) <> '')
);

CREATE INDEX IF NOT EXISTS name_import_staging_batch_status_idx
    ON data_governance.name_import_staging (batch_id, stage_status, staging_id);

COMMENT ON TABLE data_governance.name_import_staging IS
    'Quarantined import rows. They must not be promoted to names_miracle until deterministic fields, AI enrichment, review, and governed corpus evidence are complete.';
