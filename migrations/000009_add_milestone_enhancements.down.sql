-- +goose Down
-- +goose StatementBegin

-- Remove milestone enhancement columns
ALTER TABLE contract_milestones
DROP COLUMN IF EXISTS sequence_number,
DROP COLUMN IF EXISTS category,
DROP COLUMN IF EXISTS priority,
DROP COLUMN IF EXISTS critical_path,
DROP COLUMN IF EXISTS estimated_start_date,
DROP COLUMN IF EXISTS estimated_end_date,
DROP COLUMN IF EXISTS actual_start_date,
DROP COLUMN IF EXISTS actual_end_date,
DROP COLUMN IF EXISTS percentage_complete,
DROP COLUMN IF EXISTS contingency_plans;

-- Drop milestone dependency and template tables
DROP TABLE IF EXISTS milestone_dependencies;
DROP TABLE IF EXISTS milestone_templates;

-- Drop indexes
DROP INDEX IF EXISTS idx_contract_milestones_sequence_number;
DROP INDEX IF EXISTS idx_milestone_templates_name;

-- +goose StatementEnd



