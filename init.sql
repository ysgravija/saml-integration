-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Insert sample users for testing
INSERT INTO users (email, first_name, last_name, is_active) VALUES
    ('jackson@example.com', 'Jackson', 'Smith', true),
    ('test@example.com', 'Test', 'User', true),
    ('admin@example.com', 'Admin', 'User', true),
    ('inactive@example.com', 'Inactive', 'User', false)
ON CONFLICT (email) DO NOTHING; 