ALTER TABLE ride_requests
ADD COLUMN actual_fare NUMERIC,
ADD COLUMN payment_method TEXT,
ADD COLUMN rating_to_driver INT,
ADD COLUMN rating_to_customer INT,
ADD COLUMN review_comment TEXT;
