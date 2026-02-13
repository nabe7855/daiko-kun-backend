-- Add scheduled_at column to support reservations
ALTER TABLE ride_requests ADD COLUMN scheduled_at TIMESTAMPTZ;

-- Update status to include 'reserved' if needed, though 'pending' might suffice
-- for unaccepted reservations. Let's add a comment for clarity.
COMMENT ON COLUMN ride_requests.scheduled_at IS 'Scheduled time for the ride (null for immediate requests)';
