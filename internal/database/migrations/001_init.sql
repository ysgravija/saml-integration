-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index on is_active for faster filtering
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Insert sample users for testing
INSERT INTO users (email, first_name, last_name, is_active) VALUES
('test@example.com', 'Test', 'User', true),
('admin@example.com', 'Admin', 'User', true),
('inactive@example.com', 'Inactive', 'User', false)
ON CONFLICT (email) DO NOTHING;