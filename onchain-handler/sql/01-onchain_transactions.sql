CREATE TABLE onchain_transactions (
    id SERIAL PRIMARY KEY,  -- SERIAL takes care of auto-increment
    token_distribution_address VARCHAR(42) NOT NULL,
    recipient_address VARCHAR(42) NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    token_amount NUMERIC(50, 18) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,  -- 0 for pending, 1 for success, -1 for failed 
    error_message TEXT,
    tx_type VARCHAR(15) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT tx_type_check CHECK (tx_type IN ('PURCHASE', 'COMMISSION'))
);

-- Create a trigger function to update 'updated_at' column on update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger that applies to 'onchain_transactions' table
CREATE TRIGGER update_onchain_transactions_updated_at
BEFORE UPDATE ON onchain_transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

