CREATE SCHEMA IF NOT EXISTS rag_app;

SET search_path TO rag_app;

-- EXTENSIONS
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

--------------------------------------------------
-- DOCUMENTS
--------------------------------------------------
CREATE TABLE documents (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                           organization_id UUID NOT NULL,

                           filename TEXT NOT NULL,
                           content_type TEXT,

                           storage_path TEXT NOT NULL, -- S3 key

                           status TEXT NOT NULL DEFAULT 'uploaded',
    -- uploaded | processing | indexed | failed

                           size_bytes BIGINT,

                           created_at TIMESTAMP NOT NULL DEFAULT now(),
                           updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_org ON documents(organization_id);
CREATE INDEX idx_documents_status ON documents(status);

--------------------------------------------------
-- INGESTION JOBS
--------------------------------------------------
CREATE TABLE ingestion_jobs (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                                organization_id UUID NOT NULL,
                                document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

                                status TEXT NOT NULL DEFAULT 'pending',
    -- pending | processing | done | failed

                                error TEXT,

                                started_at TIMESTAMP,
                                finished_at TIMESTAMP,

                                created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_ingestion_jobs_doc ON ingestion_jobs(document_id);
CREATE INDEX idx_ingestion_jobs_status ON ingestion_jobs(status);

--------------------------------------------------
-- DOCUMENT CHUNKS
--------------------------------------------------
CREATE TABLE document_chunks (
                                 id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                                 organization_id UUID NOT NULL,
                                 document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

                                 chunk_index INT NOT NULL,
                                 content TEXT NOT NULL,

                                 metadata JSONB,

                                 created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_chunks_doc ON document_chunks(document_id);
CREATE INDEX idx_chunks_org ON document_chunks(organization_id);

--------------------------------------------------
-- RAG QUERIES
--------------------------------------------------
CREATE TABLE rag_queries (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                             organization_id UUID NOT NULL,
                             user_id UUID NOT NULL,

                             query TEXT NOT NULL,

                             created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_queries_org ON rag_queries(organization_id);

--------------------------------------------------
-- RAG ANSWERS
--------------------------------------------------
CREATE TABLE rag_answers (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                             query_id UUID NOT NULL REFERENCES rag_queries(id) ON DELETE CASCADE,

                             answer TEXT NOT NULL,

                             model TEXT,
                             latency_ms INT,

                             created_at TIMESTAMP NOT NULL DEFAULT now()
);

--------------------------------------------------
-- ANSWER SOURCES (IMPORTANT!)
--------------------------------------------------
CREATE TABLE rag_answer_sources (
                                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                                    answer_id UUID NOT NULL REFERENCES rag_answers(id) ON DELETE CASCADE,
                                    chunk_id UUID NOT NULL REFERENCES document_chunks(id) ON DELETE CASCADE,

                                    score FLOAT,

                                    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_sources_answer ON rag_answer_sources(answer_id);
CREATE INDEX idx_sources_chunk ON rag_answer_sources(chunk_id);

--------------------------------------------------
-- UPDATED_AT TRIGGER
--------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();