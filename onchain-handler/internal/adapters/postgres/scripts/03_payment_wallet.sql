CREATE TABLE IF NOT EXISTS payment_wallet (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL UNIQUE,
    in_use BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add the updated_at trigger for the payment_wallet table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_payment_wallet_updated_at'
          AND tgrelid = 'payment_wallet'::regclass
    ) THEN
        DROP TRIGGER update_payment_wallet_updated_at ON payment_wallet;
    END IF;

    CREATE TRIGGER update_payment_wallet_updated_at
    BEFORE UPDATE ON payment_wallet
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
