DO $$
BEGIN
    -- Ensure the `vendor_id` column exists
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'payment_order' AND column_name = 'vendor_id'
    ) THEN
        -- Add the column with a temporary default value
        ALTER TABLE payment_order
        ADD COLUMN vendor_id VARCHAR(33) DEFAULT '';

        -- Update existing rows to ensure the column has no NULL values
        UPDATE payment_order
        SET vendor_id = ''
        WHERE vendor_id IS NULL;

        -- Alter the column to set NOT NULL constraint
        ALTER TABLE payment_order
        ALTER COLUMN vendor_id SET NOT NULL;
    END IF;

    -- Drop any existing unique constraints on `request_id` if necessary
    IF EXISTS (
        SELECT 1 
        FROM information_schema.table_constraints 
        WHERE table_name = 'payment_order'
        AND constraint_type = 'UNIQUE'
        AND constraint_name = 'unique_request_id'
    ) THEN
        ALTER TABLE payment_order
        DROP CONSTRAINT unique_request_id;
    END IF;

    -- Add a composite unique constraint for `request_id` and `vendor_id`
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'payment_order'
        AND constraint_type = 'UNIQUE'
        AND constraint_name = 'unique_request_vendor_id'
    ) THEN
        ALTER TABLE payment_order
        ADD CONSTRAINT unique_request_vendor_id UNIQUE (request_id, vendor_id);
    END IF;
END;
$$;
