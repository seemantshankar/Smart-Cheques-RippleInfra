-- Drop views first (due to dependencies)
DROP VIEW IF EXISTS overdue_milestones_report;
DROP VIEW IF EXISTS milestone_dependency_details;
DROP VIEW IF EXISTS critical_path_milestones;
DROP VIEW IF EXISTS milestone_completion_stats;

-- Drop triggers
DROP TRIGGER IF EXISTS update_milestone_templates_updated_at ON milestone_templates;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes on existing contract_milestones table
DROP INDEX IF EXISTS idx_contract_milestones_contract_id_sequence;
DROP INDEX IF EXISTS idx_contract_milestones_category;
DROP INDEX IF EXISTS idx_contract_milestones_priority;
DROP INDEX IF EXISTS idx_contract_milestones_critical_path;
DROP INDEX IF EXISTS idx_contract_milestones_risk_level;
DROP INDEX IF EXISTS idx_contract_milestones_percentage_complete;
DROP INDEX IF EXISTS idx_contract_milestones_estimated_end_date;
DROP INDEX IF EXISTS idx_contract_milestones_actual_end_date;
DROP INDEX IF EXISTS idx_contract_milestones_criticality_score;
DROP INDEX IF EXISTS idx_contract_milestones_overdue;
DROP INDEX IF EXISTS idx_contract_milestones_contract_status;
DROP INDEX IF EXISTS idx_contract_milestones_risk_priority;
DROP INDEX IF EXISTS idx_contract_milestones_timeline;
DROP INDEX IF EXISTS idx_contract_milestones_dependencies_gin;
DROP INDEX IF EXISTS idx_contract_milestones_contingency_plans_gin;
DROP INDEX IF EXISTS idx_contract_milestones_text_search;

-- Drop indexes for milestone template shares
DROP INDEX IF EXISTS idx_milestone_template_shares_expires_at;
DROP INDEX IF EXISTS idx_milestone_template_shares_shared_by;
DROP INDEX IF EXISTS idx_milestone_template_shares_shared_with;
DROP INDEX IF EXISTS idx_milestone_template_shares_template_id;

-- Drop indexes for milestone template versions
DROP INDEX IF EXISTS idx_milestone_template_versions_created_at;
DROP INDEX IF EXISTS idx_milestone_template_versions_version;
DROP INDEX IF EXISTS idx_milestone_template_versions_template_id;

-- Drop indexes for milestone templates
DROP INDEX IF EXISTS idx_milestone_templates_text_search;
DROP INDEX IF EXISTS idx_milestone_templates_variables_gin;
DROP INDEX IF EXISTS idx_milestone_templates_created_at;
DROP INDEX IF EXISTS idx_milestone_templates_active;
DROP INDEX IF EXISTS idx_milestone_templates_category;
DROP INDEX IF EXISTS idx_milestone_templates_name;

-- Drop indexes for milestone progress history
DROP INDEX IF EXISTS idx_milestone_progress_recorded_by;
DROP INDEX IF EXISTS idx_milestone_progress_status;
DROP INDEX IF EXISTS idx_milestone_progress_recorded_at;
DROP INDEX IF EXISTS idx_milestone_progress_milestone_id;

-- Drop indexes for milestone dependencies
DROP INDEX IF EXISTS idx_milestone_dependencies_type;
DROP INDEX IF EXISTS idx_milestone_dependencies_depends_on_id;
DROP INDEX IF EXISTS idx_milestone_dependencies_milestone_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS milestone_template_shares;
DROP TABLE IF EXISTS milestone_template_versions;
DROP TABLE IF EXISTS milestone_templates;
DROP TABLE IF EXISTS milestone_progress_history;
DROP TABLE IF EXISTS milestone_dependencies;
