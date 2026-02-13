-- Create user_reports table
CREATE TABLE IF NOT EXISTS user_reports (
    id UUID PRIMARY KEY,
    ride_id UUID REFERENCES ride_requests(id),
    reporter_id UUID NOT NULL,
    reported_user_id UUID NOT NULL,
    reporter_role VARCHAR(20) NOT NULL, -- 'customer' or 'driver'
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT ('pending'), -- 'pending', 'reviewed', 'action_taken'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
