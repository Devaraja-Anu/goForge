-- name: Ping :one
-- Trivial liveness query — lets the healthcheck handler prove the DB
-- connection is actually alive (not just that it was pingable once,
-- at boot, in database.NewPool). Also the minimum sqlc needs to
-- generate a real Queries/Querier type with zero domain tables.
SELECT 1;