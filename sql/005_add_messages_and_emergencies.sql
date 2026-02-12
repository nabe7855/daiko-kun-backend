-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    ride_id UUID NOT NULL REFERENCES ride_requests(id),
    sender_id TEXT NOT NULL,
    sender_type TEXT NOT NULL, -- 'customer' or 'driver'
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create emergencies table
CREATE TABLE IF NOT EXISTS ride_emergencies (
    id UUID PRIMARY KEY,
    ride_id UUID NOT NULL REFERENCES ride_requests(id),
    reporter_id TEXT NOT NULL,
    reporter_type TEXT NOT NULL, -- 'customer' or 'driver'
    reason TEXT,
    status TEXT DEFAULT 'active', -- 'active', 'resolved'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
