-- Drop triggers
DROP TRIGGER IF EXISTS update_milestones_updated_at ON milestones;
DROP TRIGGER IF EXISTS update_contracts_updated_at ON contracts;
DROP TRIGGER IF EXISTS update_smart_cheques_updated_at ON smart_cheques;
DROP TRIGGER IF EXISTS update_enterprises_updated_at ON enterprises;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP INDEX IF EXISTS idx_milestones_status;
DROP INDEX IF EXISTS idx_milestones_smart_cheque_id;
DROP INDEX IF EXISTS idx_smart_cheques_status;
DROP INDEX IF EXISTS idx_smart_cheques_payee_id;
DROP INDEX IF EXISTS idx_smart_cheques_payer_id;
DROP INDEX IF EXISTS idx_enterprises_kyb_status;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS contracts;
DROP TABLE IF EXISTS smart_cheques;
DROP TABLE IF EXISTS enterprises;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";