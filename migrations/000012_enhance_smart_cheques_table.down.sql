-- Rollback migration for smart_cheques table enhancements

-- Drop trigger
DROP TRIGGER IF EXISTS update_smart_cheques_updated_at ON smart_cheques;

-- Drop function
DROP FUNCTION IF EXISTS update_smart_cheque_updated_at();

-- Drop view
DROP VIEW IF EXISTS smart_cheque_analytics;

-- Drop indexes
DROP INDEX IF EXISTS idx_smart_cheques_payer_id;
DROP INDEX IF EXISTS idx_smart_cheques_payee_id;
DROP INDEX IF EXISTS idx_smart_cheques_status;
DROP INDEX IF EXISTS idx_smart_cheques_contract_hash;
DROP INDEX IF EXISTS idx_smart_cheques_created_at;

-- Note: We're not dropping the milestones column or constraints as they might be needed
-- In a real rollback scenario, you might want to handle this differently based on your requirements