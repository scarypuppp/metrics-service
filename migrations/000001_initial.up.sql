CREATE TABLE metrics
(
    id    VARCHAR NOT NULL,
    mtype VARCHAR NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    hash  VARCHAR NOT NULL
);

CREATE UNIQUE INDEX idx_metrics_id ON metrics (id, mtype);