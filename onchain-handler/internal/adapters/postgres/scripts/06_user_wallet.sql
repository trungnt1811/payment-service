CREATE TABLE IF NOT EXISTS user_wallet (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    address VARCHAR(42) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add the updated_at trigger for the user_wallet table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_user_wallet_updated_at'
          AND tgrelid = 'user_wallet'::regclass
    ) THEN
        DROP TRIGGER update_user_wallet_updated_at ON user_wallet;
    END IF;

    CREATE TRIGGER update_user_wallet_updated_at
    BEFORE UPDATE ON user_wallet
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
