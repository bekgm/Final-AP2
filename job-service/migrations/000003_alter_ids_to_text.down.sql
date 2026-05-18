-- migrations/000003_alter_ids_to_text.down.sql
-- Revert identifiers back to UUID.

ALTER TABLE applications
    ALTER COLUMN freelancer_id TYPE UUID USING freelancer_id::uuid;

ALTER TABLE jobs
    ALTER COLUMN client_id TYPE UUID USING client_id::uuid;

