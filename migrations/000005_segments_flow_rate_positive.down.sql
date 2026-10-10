ALTER TABLE segments DROP CONSTRAINT segments_flow_rate_positive;
ALTER TABLE segments ALTER COLUMN flow_rate SET DEFAULT 0;