CREATE TABLE lock_event (
    id BIGSERIAL PRIMARY KEY,
    user_address CHAR(42) NOT NULL,
    lock_id BIGINT NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL UNIQUE,
    amount DECIMAL(50, 18),
    current_balance DECIMAL(50, 18),
    lock_action VARCHAR(10) NOT NULL, --DEPOSIT or WITHDRAW
    status SMALLINT NOT NULL DEFAULT 0,  -- 0 for pending, 1 for success, -1 for failed 
    lock_duration BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_duration TIMESTAMP,
    CONSTRAINT lock_event_transaction_hash_unique UNIQUE (transaction_hash),
    CONSTRAINT lock_action_check CHECK (lock_action IN ('DEPOSIT', 'WITHDRAW'))
);

CREATE INDEX lock_event_user_address_status_lock_id_idx ON lock_event (user_address, status, lock_id);
CREATE INDEX lock_event_lock_id_created_at_idx ON lock_event (lock_id, created_at);
CREATE INDEX lock_event_lock_id_lock_action_idx ON lock_event (lock_id, lock_action);
CREATE INDEX lock_event_user_address_idx ON lock_event (user_address);

CREATE TRIGGER update_lock_event_updated_at
BEFORE UPDATE ON lock_event
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();