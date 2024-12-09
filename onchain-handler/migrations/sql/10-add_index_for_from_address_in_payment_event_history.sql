-- Add index for `from_address` column
CREATE INDEX IF NOT EXISTS payment_event_history_from_address_idx 
ON payment_event_history (from_address);