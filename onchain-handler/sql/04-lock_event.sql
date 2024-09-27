CREATE TABLE lock_event (
    id BIGSERIAL PRIMARY KEY,
    user_address CHAR(42) NOT NULL,
    lock_id BIGINT NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL UNIQUE,
    amount DECIMAL(50, 18),
    lock_action VARCHAR(10) NOT NULL, --DEPOSIT or WITHDRAW
    status SMALLINT NOT NULL DEFAULT 0,  -- 0 for pending, 1 for success, -1 for failed 
    lock_duration BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_duration TIMESTAMP NOT NULL,
    CONSTRAINT lock_event_transaction_hash_unique UNIQUE (transaction_hash),
    CONSTRAINT lock_action_check CHECK (lock_action IN ('DEPOSIT', 'WITHDRAW'))
);

CREATE TRIGGER update_lock_event_updated_at
BEFORE UPDATE ON lock_event
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();