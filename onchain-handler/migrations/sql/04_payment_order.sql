DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM('PENDING', 'PROCESSING', 'SUCCESS', 'PARTIAL', 'EXPIRED', 'FAILED');
    END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS payment_order (
    id SERIAL PRIMARY KEY,
    request_id VARCHAR(255) NOT NULL UNIQUE,
    wallet_id INT NOT NULL REFERENCES payment_wallet(id) ON DELETE CASCADE, -- foreign key constraint
    block_height INT NOT NULL,
    amount NUMERIC(30, 18) NOT NULL,
    transferred NUMERIC(30, 18) NOT NULL DEFAULT 0,
    symbol VARCHAR(10) NOT NULL,
    network VARCHAR(20) NOT NULL,
    status order_status NOT NULL DEFAULT 'PENDING',
    webhook_url VARCHAR(255) NOT NULL,
    succeeded_at TIMESTAMP WITH TIME ZONE,
    expired_time TIMESTAMP WITH TIME ZONE NOT NULL, 
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create a trigger for the payment_order table to update 'updated_at'
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_payment_order_updated_at'
          AND tgrelid = 'payment_order'::regclass
    ) THEN
        DROP TRIGGER update_payment_order_updated_at ON payment_order;
    END IF;

    CREATE TRIGGER update_payment_order_updated_at
    BEFORE UPDATE ON payment_order
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;

