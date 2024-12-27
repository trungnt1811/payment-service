-- Indexes for filtering
CREATE INDEX IF NOT EXISTS payment_order_status_idx ON payment_order (status);
CREATE INDEX IF NOT EXISTS payment_order_network_idx ON payment_order (network);
CREATE INDEX IF NOT EXISTS payment_order_created_at_idx ON payment_order (created_at);
CREATE INDEX IF NOT EXISTS payment_order_succeeded_at_idx ON payment_order (succeeded_at);
CREATE INDEX IF NOT EXISTS payment_order_expired_time_idx ON payment_order (expired_time);

-- Composite indexes for combined filtering and sorting
CREATE INDEX IF NOT EXISTS payment_order_status_succeeded_at_idx ON payment_order (status, succeeded_at);
CREATE INDEX IF NOT EXISTS payment_order_status_expired_time_idx ON payment_order (status, expired_time);
CREATE INDEX IF NOT EXISTS payment_order_network_created_at_idx ON payment_order (network, created_at);
CREATE INDEX IF NOT EXISTS payment_order_network_succeeded_at_idx ON payment_order (network, succeeded_at);
