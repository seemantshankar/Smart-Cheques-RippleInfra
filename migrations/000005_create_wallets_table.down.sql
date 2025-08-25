DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_wallets_enterprise_network_active;
DROP INDEX IF EXISTS idx_wallets_network_type;
DROP INDEX IF EXISTS idx_wallets_is_whitelisted;
DROP INDEX IF EXISTS idx_wallets_status;
DROP INDEX IF EXISTS idx_wallets_address;
DROP INDEX IF EXISTS idx_wallets_enterprise_id;
DROP TABLE IF EXISTS wallets;