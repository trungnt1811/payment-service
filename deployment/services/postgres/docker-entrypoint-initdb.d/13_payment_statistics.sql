DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'granularity_type') THEN
        CREATE TYPE granularity_type AS ENUM ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY');
    END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS payment_statistics (
    id SERIAL PRIMARY KEY,
    granularity granularity_type NOT NULL, -- Enum for granularity
    period_start TIMESTAMP WITH TIME ZONE NOT NULL, -- Start of the period (e.g., start of day/week/month)
    total_orders BIGINT NOT NULL DEFAULT 0, -- Total number of orders
    total_amount NUMERIC(30, 18) NOT NULL DEFAULT 0, -- Total amount in the period
    total_transferred NUMERIC(30, 18) NOT NULL DEFAULT 0, -- Total transferred in the period
    symbol VARCHAR(10) NOT NULL,
    vendor_id VARCHAR(33) DEFAULT '',
    is_aggregated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (granularity, period_start, symbol, vendor_id) -- Ensure no duplicate entries for the same period
);

-- Composite index on granularity, period_start, and vendor_id
CREATE INDEX IF NOT EXISTS idx_payment_statistics_granularity_period_vendor
ON payment_statistics (granularity, period_start, vendor_id);

-- Partial index for unaggregated records
CREATE INDEX IF NOT EXISTS idx_payment_statistics_unaggregated
ON payment_statistics (granularity, period_start, vendor_id)
WHERE is_aggregated = FALSE;

-- Add the updated_at trigger for the payment_statistics table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_payment_statistics_updated_at'
          AND tgrelid = 'payment_statistics'::regclass
    ) THEN
        DROP TRIGGER update_payment_statistics_updated_at ON payment_statistics;
    END IF;

    CREATE TRIGGER update_payment_statistics_updated_at
    BEFORE UPDATE ON payment_statistics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;

