-- migrations/000001_create_jobs.up.sql
CREATE TABLE IF NOT EXISTS jobs (
    id          UUID PRIMARY KEY,
    client_id   UUID        NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    budget      NUMERIC(12, 2) NOT NULL DEFAULT 0,
    status      VARCHAR(50)  NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_jobs_client_id ON jobs(client_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status    ON jobs(status);
