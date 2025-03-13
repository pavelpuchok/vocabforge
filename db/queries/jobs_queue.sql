-- name: CreateJob :exec
INSERT INTO jobs_queue (group_name, item, created_at) VALUES (?, ?, ?);

-- name: GetOldestJobByGroup :one
SELECT id,item FROM jobs_queue 
WHERE group_name=?
ORDER BY created_at ASC
LIMIT 1;

-- name: DeleteJob :exec
DELETE FROM jobs_queue WHERE id = ?;
