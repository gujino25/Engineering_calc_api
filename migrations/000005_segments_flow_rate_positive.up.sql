DELETE FROM segments WHERE flow_rate = 0;

ALTER TABLE segments ALTER COLUMN flow_rate DROP DEFAULT;
ALTER TABLE segments ADD CONSTRAINT segments_flow_rate_positive CHECK (flow_rate > 0);