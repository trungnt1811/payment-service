CREATE TABLE IF NOT EXISTS user_wallet (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    address VARCHAR(42) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX user_wallet_user_id_idx ON user_wallet (user_id);
CREATE UNIQUE INDEX user_wallet_address_idx ON user_wallet (address);