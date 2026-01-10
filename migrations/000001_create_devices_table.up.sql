-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    brand VARCHAR(50) NOT NULL,
    state VARCHAR(20) NOT NULL CHECK (state IN ('active', 'in-use', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX idx_devices_brand ON devices(brand);
CREATE INDEX idx_devices_state ON devices(state);
CREATE INDEX idx_devices_created_at ON devices(created_at DESC);
