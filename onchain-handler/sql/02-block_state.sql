CREATE TABLE block_state (
    id SERIAL PRIMARY KEY,  
    last_block BIGINT NOT NULL       -- Stores the last processed block number
);
