-- migrations/000002_create_applications.up.sql
CREATE TABLE IF NOT EXISTS applications (
    id            UUID PRIMARY KEY,
    job_id        UUID        NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    freelancer_id UUID        NOT NULL,
    cover_letter  TEXT,
    status        VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_applications_job_id        ON applications(job_id);
CREATE INDEX IF NOT EXISTS idx_applications_freelancer_id ON applications(freelancer_id);
