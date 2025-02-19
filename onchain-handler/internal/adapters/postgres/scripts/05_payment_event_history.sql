CREATE TABLE IF NOT EXISTS payment_event_history (
    id SERIAL PRIMARY KEY,
    payment_order_id INT NOT NULL REFERENCES payment_order(id) ON DELETE CASCADE,
    transaction_hash VARCHAR(66) NOT NULL UNIQUE,
    from_address VARCHAR(42) NOT NULL,                 
    to_address VARCHAR(42) NOT NULL,                    
    contract_address VARCHAR(42) NOT NULL,             
    token_symbol VARCHAR(10) NOT NULL,     
    network VARCHAR(20) NOT NULL,            
    amount NUMERIC(30, 18) NOT NULL,                   
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS payment_event_history_payment_order_id_idx ON payment_event_history (payment_order_id);
CREATE INDEX IF NOT EXISTS payment_event_history_from_address_idx ON payment_event_history (from_address, payment_order_id);

-- Add the updated_at trigger for the payment_event_history table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_payment_event_history_updated_at'
          AND tgrelid = 'payment_event_history'::regclass
    ) THEN
        DROP TRIGGER update_payment_event_history_updated_at ON payment_event_history;
    END IF;

    CREATE TRIGGER update_payment_event_history_updated_at
    BEFORE UPDATE ON payment_event_history
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
