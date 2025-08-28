-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_contract_milestones_contract_id;
DROP INDEX IF EXISTS idx_contract_milestones_milestone_id;
DROP INDEX IF EXISTS idx_contracts_status;
DROP INDEX IF EXISTS idx_contracts_contract_type;

-- Drop contract_milestones table
DROP TABLE IF EXISTS contract_milestones;

-- Remove added columns from contracts table
ALTER TABLE contracts
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS contract_type,
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS parent_contract_id,
DROP COLUMN IF EXISTS original_filename,
DROP COLUMN IF EXISTS file_size,
DROP COLUMN IF EXISTS mime_type,
DROP COLUMN IF EXISTS expiration_date,
DROP COLUMN IF EXISTS renewal_terms;

-- +goose StatementEnd
