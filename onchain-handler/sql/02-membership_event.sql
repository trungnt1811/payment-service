CREATE TABLE membership_event (
    id BIGSERIAL PRIMARY KEY,
    user_address VARCHAR(50) NOT NULL,
    order_id BIGINT NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL UNIQUE,
    amount DECIMAL(50, 18),
    status SMALLINT NOT NULL DEFAULT 0,  -- 0 for pending, 1 for success, -1 for failed 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_duration TIMESTAMP NOT NULL, -- The expiry date/time of the membership
    CONSTRAINT membership_event_transaction_hash_unique UNIQUE (transaction_hash)
);

CREATE INDEX membership_event_order_id_idx ON membership_event (order_id);
CREATE INDEX membership_event_status_end_duration_idx ON membership_event (status, end_duration);

-- Create a trigger to update 'updated_at' column on update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_membership_event_updated_at
BEFORE UPDATE ON membership_event
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();