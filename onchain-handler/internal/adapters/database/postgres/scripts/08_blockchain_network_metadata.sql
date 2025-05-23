CREATE TABLE IF NOT EXISTS blockchain_network_metadata (
    id SERIAL PRIMARY KEY,
    alias VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL UNIQUE,
    icon_base64 TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

DO $$
BEGIN
    INSERT INTO blockchain_network_metadata (alias, name, icon_base64)
    VALUES 
        (
            'BSC', 
            'BNB Smart Chain (BEP20)',
            'data:image/png;base64,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'
        ),
        (
            'AVAX C-Chain', 
            'Avalanche C-Chain',
            'data:image/png;base64,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'
        )
    ON CONFLICT (alias) DO NOTHING;
END;    
$$;

-- Add the updated_at trigger for the blockchain_network_metadata table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_blockchain_network_metadata_updated_at'
          AND tgrelid = 'blockchain_network_metadata'::regclass
    ) THEN
        DROP TRIGGER update_blockchain_network_metadata_updated_at ON blockchain_network_metadata;
    END IF;

    CREATE TRIGGER update_blockchain_network_metadata_updated_at
    BEFORE UPDATE ON blockchain_network_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;

