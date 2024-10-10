CREATE TABLE IF NOT EXISTS block_state (
    id SERIAL PRIMARY KEY,  
    last_block BIGINT NOT NULL       -- Stores the last processed block number
);
