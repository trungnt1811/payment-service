DO $$
BEGIN
    -- Drop the composite unique constraint on `request_id` and `vendor_id`
    IF EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'payment_order'
        AND constraint_type = 'UNIQUE'
        AND constraint_name = 'unique_request_vendor_id'
    ) THEN
        ALTER TABLE payment_order
        DROP CONSTRAINT unique_request_vendor_id;
    END IF;

    -- Restore the unique constraint on `request_id` (if it existed previously)
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'payment_order'
        AND constraint_type = 'UNIQUE'
        AND constraint_name = 'unique_request_id'
    ) THEN
        ALTER TABLE payment_order
        ADD CONSTRAINT unique_request_id UNIQUE (request_id);
    END IF;
END;
$$;
