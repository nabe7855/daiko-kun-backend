CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    phone_number TEXT NOT NULL UNIQUE,
    email TEXT UNIQUE,
    social_id TEXT, -- For Google/Facebook login
    social_provider TEXT, -- 'google', 'facebook'
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Update ride_requests to link to customers table
-- (Optional, but good for data integrity. Let's keep it simple for now)
