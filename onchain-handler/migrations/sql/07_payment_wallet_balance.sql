CREATE TABLE IF NOT EXISTS payment_wallet_balance (
    id SERIAL PRIMARY KEY,
    wallet_id INT NOT NULL REFERENCES payment_wallet(id) ON DELETE CASCADE,
    network VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    balance NUMERIC(30, 18) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS payment_wallet_balance_wallet_id_network_symbol_idx ON payment_wallet_balance (wallet_id, network, symbol);