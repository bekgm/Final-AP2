-- migrations/000003_alter_ids_to_text.up.sql
-- The platform may use non-UUID identifiers for clients/freelancers.
-- Store them as TEXT to avoid PostgreSQL UUID cast errors.

ALTER TABLE jobs
    ALTER COLUMN client_id TYPE TEXT USING client_id::text;

ALTER TABLE applications
    ALTER COLUMN freelancer_id TYPE TEXT USING freelancer_id::text;

