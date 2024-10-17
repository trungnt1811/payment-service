CREATE TYPE order_status AS ENUM('PENDING', 'SUCCESS', 'PARTIAL', 'EXPIRED', 'FAILED');

CREATE TABLE IF NOT EXISTS payment_order (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    wallet_id INT NOT NULL REFERENCES payment_wallet(id) ON DELETE CASCADE, -- foreign key constraint
    block_height INT NOT NULL,
    amount NUMERIC(30, 18) NOT NULL,
    transferred NUMERIC(30, 18) NOT NULL DEFAULT 0,
    status order_status NOT NULL DEFAULT 'PENDING',
    expired_time TIMESTAMP NOT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX payment_order_status_idx ON payment_order (status);
CREATE INDEX payment_order_expired_time_idx ON payment_order (expired_time);
CREATE INDEX payment_order_block_height_idx ON payment_order (block_height);
CREATE INDEX payment_order_status_expired_time_idx ON payment_order (status, expired_time);