-- Add new type 'INTERNAL_TRANSFER' to the 'transfer_type' enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_enum 
        WHERE enumlabel = 'INTERNAL_TRANSFER' 
          AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'transfer_type')
    ) THEN
        ALTER TYPE transfer_type ADD VALUE 'INTERNAL_TRANSFER';
    END IF;
END;
$$;
