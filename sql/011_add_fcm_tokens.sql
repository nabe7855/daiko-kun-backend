-- Add fcm_token column to customers and drivers tables
ALTER TABLE customers ADD COLUMN IF NOT EXISTS fcm_token TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS fcm_token TEXT;
