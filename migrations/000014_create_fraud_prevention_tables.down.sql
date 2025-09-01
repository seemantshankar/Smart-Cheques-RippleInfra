-- Drop fraud prevention tables
-- Migration: 000014_create_fraud_prevention_tables.down.sql

-- Drop views
DROP VIEW IF EXISTS fraud_alert_summary;
DROP VIEW IF EXISTS restricted_accounts;
DROP VIEW IF EXISTS high_priority_fraud_cases;
DROP VIEW IF EXISTS active_fraud_rules;

-- Drop triggers
DROP TRIGGER IF EXISTS update_account_fraud_status_updated_at ON account_fraud_status;
DROP TRIGGER IF EXISTS update_fraud_cases_updated_at ON fraud_cases;
DROP TRIGGER IF EXISTS update_fraud_rules_updated_at ON fraud_rules;
DROP TRIGGER IF EXISTS update_fraud_alerts_updated_at ON fraud_alerts;

-- Drop function
DROP FUNCTION IF EXISTS update_fraud_prevention_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_fraud_alerts_title_desc_fts;
DROP INDEX IF EXISTS idx_fraud_rules_name_desc_fts;
DROP INDEX IF EXISTS idx_fraud_cases_title_desc_fts;

DROP INDEX IF EXISTS idx_account_fraud_status_limits_gin;
DROP INDEX IF EXISTS idx_account_fraud_status_restrictions_gin;
DROP INDEX IF EXISTS idx_fraud_cases_resolution_gin;
DROP INDEX IF EXISTS idx_fraud_rules_thresholds_gin;
DROP INDEX IF EXISTS idx_fraud_rules_conditions_gin;
DROP INDEX IF EXISTS idx_fraud_alerts_evidence_gin;

DROP INDEX IF EXISTS idx_account_fraud_status_next_review_date;
DROP INDEX IF EXISTS idx_account_fraud_status_risk_level;
DROP INDEX IF EXISTS idx_account_fraud_status_status;

DROP INDEX IF EXISTS idx_fraud_cases_case_number;
DROP INDEX IF EXISTS idx_fraud_cases_opened_at;
DROP INDEX IF EXISTS idx_fraud_cases_category;
DROP INDEX IF EXISTS idx_fraud_cases_priority;
DROP INDEX IF EXISTS idx_fraud_cases_status;
DROP INDEX IF EXISTS idx_fraud_cases_enterprise_id;

DROP INDEX IF EXISTS idx_fraud_rules_active;
DROP INDEX IF EXISTS idx_fraud_rules_created_by;
DROP INDEX IF EXISTS idx_fraud_rules_effective_at;
DROP INDEX IF EXISTS idx_fraud_rules_category;
DROP INDEX IF EXISTS idx_fraud_rules_status;

DROP INDEX IF EXISTS idx_fraud_alerts_rule_id;
DROP INDEX IF EXISTS idx_fraud_alerts_case_id;
DROP INDEX IF EXISTS idx_fraud_alerts_alert_type;
DROP INDEX IF EXISTS idx_fraud_alerts_detected_at;
DROP INDEX IF EXISTS idx_fraud_alerts_severity;
DROP INDEX IF EXISTS idx_fraud_alerts_status;
DROP INDEX IF EXISTS idx_fraud_alerts_enterprise_id;

-- Drop tables
DROP TABLE IF EXISTS account_fraud_status;
DROP TABLE IF EXISTS fraud_cases;
DROP TABLE IF EXISTS fraud_rules;
DROP TABLE IF EXISTS fraud_alerts;
