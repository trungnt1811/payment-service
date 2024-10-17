CREATE TABLE IF NOT EXISTS block_state (
    id SERIAL PRIMARY KEY,  
    latest_block BIGINT NOT NULL, -- Catchup the latest block from chain
    last_processed_block BIGINT NOT NULL -- Stores the last processed block number       
);
