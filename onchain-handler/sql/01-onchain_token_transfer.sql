CREATE TABLE IF NOT EXISTS onchain_token_transfer (
    id SERIAL PRIMARY KEY,  -- SERIAL takes care of auto-increment
    request_id VARCHAR(15) NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    token_amount NUMERIC(50, 18) NOT NULL,
    fee NUMERIC(50, 18) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    from_pool_name VARCHAR(15) NOT NULL,
    status BOOLEAN NOT NULL DEFAULT False,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX onchain_token_transfer_transaction_hash_idx ON onchain_token_transfer (transaction_hash);
CREATE INDEX onchain_token_transfer_from_address_idx ON onchain_token_transfer (from_address);
CREATE INDEX onchain_token_transfer_to_address_idx ON onchain_token_transfer (to_address);
CREATE INDEX onchain_token_transfer_symbol_idx ON onchain_token_transfer (symbol);
CREATE INDEX onchain_token_transfer_from_pool_name_idx ON onchain_token_transfer (from_pool_name);
CREATE INDEX onchain_token_transfer_created_at_idx ON onchain_token_transfer (created_at);

-- Create a trigger function to update 'updated_at' column on update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger that applies to 'onchain_token_transfer' table
CREATE TRIGGER update_onchain_token_transfer_updated_at
BEFORE UPDATE ON onchain_token_transfer
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

