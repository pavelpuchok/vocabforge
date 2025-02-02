-- name: AddWord :one
INSERT INTO vocab_words ( preply_id, spelling, definition, lexical_category, lang, translation_ru,  user_id, added_at, answered_count, viewed_count) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0) RETURNING *;

-- name: GetWordByID :one
SELECT * FROM vocab_words WHERE id = ?;

-- name: GetOldestUnseenWord :one
SELECT * FROM vocab_words WHERE user_id = ? AND viewed_count = 0 ORDER BY added_at DESC LIMIT 1;

-- name: CountLearningWords :one
SELECT COUNT(*) FROM vocab_words WHERE user_id = ? AND viewed_count > 0 AND learned_at IS NULL;

-- name: GetLearningWord :one
SELECT * FROM vocab_words WHERE user_id = ? AND viewed_count > 0 AND learned_at IS NULL ORDER BY last_showed_at ASC LIMIT 1;

-- name: IncrementWordViewedCountByID :exec
UPDATE vocab_words SET viewed_count = viewed_count + 1, last_showed_at = ? WHERE id = ?;

-- name: IncrementWordAnsweredCount :one
UPDATE vocab_words SET viewed_count = viewed_count + 1, last_showed_at = ? WHERE id = ? RETURNING answered_count;

-- name: ResetWordAnsweredCount :exec
UPDATE vocab_words SET viewed_count = 0, last_showed_at = ? WHERE id = ?;

-- name: SetWordLearned :exec
UPDATE vocab_words SET learned_at = ? WHERE id = ? ;

-- name: CreateExercise :one
INSERT INTO vocab_words_exercises (word_id, user_id, question, answer, telegram_msg_id, answered, created_at) 
VALUES(?,?,?,?,?,false,?) RETURNING *;

-- name: GetUnansweredExerciseByUserID :one
SELECT * FROM vocab_words_exercises WHERE user_id=? AND answered=false ORDER BY telegram_msg_id ASC LIMIT 1;

-- name: GetExerciseByTelegramIDAndUserID :one
SELECT * FROM vocab_words_exercises WHERE user_id=? AND telegram_msg_id=? ;


-- name: SetExerciseAnswer :one
UPDATE  vocab_words_exercises SET 
answered = ?,
answered_correctly= ?,
answered_at = ?
WHERE user_id = ? AND telegram_msg_id = ? RETURNING *;
