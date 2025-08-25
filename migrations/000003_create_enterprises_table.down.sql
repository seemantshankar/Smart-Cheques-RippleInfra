-- Drop triggers
DROP TRIGGER IF EXISTS update_enterprises_updated_at ON enterprises;

-- Drop indexes
DROP INDEX IF EXISTS idx_enterprise_documents_status;
DROP INDEX IF EXISTS idx_enterprise_documents_document_type;
DROP INDEX IF EXISTS idx_enterprise_documents_enterprise_id;

DROP INDEX IF EXISTS idx_authorized_representatives_email;
DROP INDEX IF EXISTS idx_authorized_representatives_enterprise_id;

DROP INDEX IF EXISTS idx_enterprises_created_at;
DROP INDEX IF EXISTS idx_enterprises_jurisdiction;
DROP INDEX IF EXISTS idx_enterprises_compliance_status;
DROP INDEX IF EXISTS idx_enterprises_kyb_status;
DROP INDEX IF EXISTS idx_enterprises_registration_number;

-- Drop tables
DROP TABLE IF EXISTS enterprise_documents;
DROP TABLE IF EXISTS authorized_representatives;
DROP TABLE IF EXISTS enterprises;