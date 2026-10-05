-- Native bigint parameters and results retain precision across the wire.
-- name: EchoBigInts :one
SELECT
    sqlc.arg(value)::bigint AS value,
    ARRAY[sqlc.arg(value)::bigint]::bigint[] AS values;
