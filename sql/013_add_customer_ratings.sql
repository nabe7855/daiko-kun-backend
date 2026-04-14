-- Add rating columns to customers table
ALTER TABLE customers
ADD COLUMN average_rating DECIMAL DEFAULT 0,
ADD COLUMN rating_count INT DEFAULT 0;
