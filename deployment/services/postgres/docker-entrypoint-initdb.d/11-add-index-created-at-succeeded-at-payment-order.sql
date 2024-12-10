-- Composite index for filtering by status and sorting by succeeded_at
CREATE INDEX IF NOT EXISTS payment_order_status_succeeded_at_idx ON payment_order (status, succeeded_at);

-- Composite index for filtering by network and sorting by created_at
CREATE INDEX IF NOT EXISTS payment_order_network_created_at_idx ON payment_order (network, created_at);
