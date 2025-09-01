-- Drop categorization rules tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_categorization_rules_updated_at ON categorization_rules;
DROP TRIGGER IF EXISTS update_categorization_rule_groups_updated_at ON categorization_rule_groups;
DROP TRIGGER IF EXISTS update_categorization_rule_templates_updated_at ON categorization_rule_templates;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS categorization_rule_performance;
DROP TABLE IF EXISTS categorization_rule_templates;
DROP TABLE IF EXISTS categorization_rule_groups;
DROP TABLE IF EXISTS categorization_rules;
