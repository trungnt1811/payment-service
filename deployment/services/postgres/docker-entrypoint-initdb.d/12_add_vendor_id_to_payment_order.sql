-- Add `vendor_id` column to the `payment_order` table
DO $$
BEGIN
    -- Check if the column `vendor_id` already exists
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

        -- Alter the column to set it as NOT NULL
        ALTER TABLE payment_order
        ALTER COLUMN vendor_id SET NOT NULL;
    END IF;
END;
$$;
