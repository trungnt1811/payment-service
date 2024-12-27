SET TIMEZONE TO 'UTC';

-- Check if the enum type exists, and create it if it doesn't
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transfer_type') THEN
        CREATE TYPE transfer_type AS ENUM ('TRANSFER', 'WITHDRAW', 'DEPOSIT');
    END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS onchain_token_transfer (
    id SERIAL PRIMARY KEY,  -- SERIAL takes care of auto-increment
    request_id VARCHAR(255) NOT NULL,
    network VARCHAR(20) NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    token_amount NUMERIC(30, 18) NOT NULL,
    fee NUMERIC(30, 18) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    status BOOLEAN NOT NULL DEFAULT FALSE,
    type transfer_type NOT NULL DEFAULT 'TRANSFER',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes if they do not exist
CREATE INDEX IF NOT EXISTS onchain_token_transfer_request_id_idx ON onchain_token_transfer (request_id);
CREATE INDEX IF NOT EXISTS onchain_token_transfer_created_at_idx ON onchain_token_transfer (created_at);

-- Create a trigger function to update 'updated_at' column on update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists, then create it
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_onchain_token_transfer_updated_at'
          AND tgrelid = 'onchain_token_transfer'::regclass
    ) THEN
        DROP TRIGGER update_onchain_token_transfer_updated_at ON onchain_token_transfer;
    END IF;

    CREATE TRIGGER update_onchain_token_transfer_updated_at
    BEFORE UPDATE ON onchain_token_transfer
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;

