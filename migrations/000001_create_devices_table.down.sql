-- Drop indexes
DROP INDEX IF EXISTS idx_devices_created_at;
DROP INDEX IF EXISTS idx_devices_state;
DROP INDEX IF EXISTS idx_devices_brand;

-- Drop table
DROP TABLE IF EXISTS devices;
