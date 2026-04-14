-- Add image_url to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS image_url TEXT;

-- Make content optional if needed (already allows empty string but let's be explicitly clear)
-- Actually NOT NULL is fine as long as we send empty string.
