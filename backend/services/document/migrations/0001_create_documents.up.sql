CREATE TABLE IF NOT EXISTS documents (
    id              uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    file_name       text NOT NULL,
    content_type    text NOT NULL,
    status          text NOT NULL
        CHECK (status IN ('pending_upload', 'uploaded', 'processing', 'indexed', 'failed')),
    object_key      text NOT NULL UNIQUE,
    size_bytes      bigint NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    error_message   text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS documents_org_created_idx
    ON documents (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS documents_org_status_idx
    ON documents (organization_id, status);

CREATE INDEX IF NOT EXISTS documents_object_key_idx
    ON documents (object_key);
