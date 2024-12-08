-- Add the upcoming_block_height column
DO $$
BEGIN
    -- Check if the column exists before attempting to add it
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'payment_order' AND column_name = 'upcoming_block_height'
    ) THEN
        ALTER TABLE payment_order
        ADD COLUMN upcoming_block_height INT NOT NULL DEFAULT 0;
    END IF;
END;
$$;
