-- Migration Down: Drop assets and balances tables
-- Description: Remove tables for asset management

-- Drop views first
DROP VIEW IF EXISTS enterprise_balance_summary;

-- Drop triggers
DROP TRIGGER IF EXISTS update_asset_transactions_updated_at ON asset_transactions;
DROP TRIGGER IF EXISTS update_enterprise_balances_updated_at ON enterprise_balances;
DROP TRIGGER IF EXISTS update_supported_assets_updated_at ON supported_assets;

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS asset_transactions;
DROP TABLE IF EXISTS enterprise_balances;
DROP TABLE IF EXISTS supported_assets;