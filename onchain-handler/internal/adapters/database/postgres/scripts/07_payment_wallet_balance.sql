CREATE TABLE IF NOT EXISTS payment_wallet_balance (
    id SERIAL PRIMARY KEY,
    wallet_id INT NOT NULL REFERENCES payment_wallet(id) ON DELETE CASCADE,
    network VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    balance NUMERIC(30, 18) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS payment_wallet_balance_wallet_id_network_symbol_idx ON payment_wallet_balance (wallet_id, network, symbol);

-- Add the updated_at trigger for the payment_wallet_balance table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_payment_wallet_balance_updated_at'
          AND tgrelid = 'payment_wallet_balance'::regclass
    ) THEN
        DROP TRIGGER update_payment_wallet_balance_updated_at ON payment_wallet_balance;
    END IF;

    CREATE TRIGGER update_payment_wallet_balance_updated_at
    BEFORE UPDATE ON payment_wallet_balance
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
