CREATE TYPE order_status AS ENUM('PENDING', 'SUCCESS', 'PARTIAL', 'EXPIRED', 'FAILED');

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
    succeeded_at TIMESTAMP WITH TIME ZONE,
    expired_time TIMESTAMP WITH TIME ZONE NOT NULL, 
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX payment_order_request_id_idx ON payment_order (request_id);
CREATE INDEX payment_order_expired_time_idx ON payment_order (expired_time);
CREATE INDEX payment_order_status_expired_time_idx ON payment_order (status, expired_time);