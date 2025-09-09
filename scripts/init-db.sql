-- LightChain L2 Database Initialization

-- Create schemas for different node types
CREATE SCHEMA IF NOT EXISTS validator;
CREATE SCHEMA IF NOT EXISTS sequencer;
CREATE SCHEMA IF NOT EXISTS archive;

-- Create users with appropriate permissions
CREATE USER validator_user WITH PASSWORD 'validator123';
CREATE USER sequencer_user WITH PASSWORD 'sequencer123';
CREATE USER archive_user WITH PASSWORD 'archive123';

-- Grant permissions
GRANT ALL ON SCHEMA validator TO validator_user;
GRANT ALL ON SCHEMA sequencer TO sequencer_user;
GRANT ALL ON SCHEMA archive TO archive_user;

-- Create basic tables for state management
CREATE TABLE IF NOT EXISTS validator.blocks (
    number BIGINT PRIMARY KEY,
    hash VARCHAR(66) NOT NULL,
    parent_hash VARCHAR(66) NOT NULL,
    timestamp BIGINT NOT NULL,
    data JSONB
);

CREATE TABLE IF NOT EXISTS sequencer.transactions (
    hash VARCHAR(66) PRIMARY KEY,
    block_number BIGINT,
    tx_index INTEGER,
    from_address VARCHAR(42),
    to_address VARCHAR(42),
    value NUMERIC,
    data JSONB,
    timestamp BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS archive.full_state (
    key VARCHAR(66) PRIMARY KEY,
    value BYTEA,
    block_number BIGINT NOT NULL,
    timestamp BIGINT NOT NULL
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON validator.blocks(timestamp);
CREATE INDEX IF NOT EXISTS idx_transactions_block ON sequencer.transactions(block_number);
CREATE INDEX IF NOT EXISTS idx_full_state_block ON archive.full_state(block_number);

-- Insert genesis data
INSERT INTO validator.blocks (number, hash, parent_hash, timestamp, data) 
VALUES (0, '0x0000000000000000000000000000000000000000000000000000000000000000', 
        '0x0000000000000000000000000000000000000000000000000000000000000000', 
        EXTRACT(EPOCH FROM NOW())::BIGINT, '{"genesis": true}')
ON CONFLICT (number) DO NOTHING;
