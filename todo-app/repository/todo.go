package repository

import (
	"database/sql"
	"todo-app/models"
)

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) FindByUserID(userID int) ([]models.Todo, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, title, completed, created_at, updated_at
		 FROM todos WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []models.Todo{}
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *TodoRepository) Create(userID int, title string) (models.Todo, error) {
	var todo models.Todo
	err := r.db.QueryRow(
		`INSERT INTO todos (user_id, title) VALUES ($1, $2)
		 RETURNING id, user_id, title, completed, created_at, updated_at`,
		userID, title,
	).Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	return todo, err
}

// Delete 逻辑删除，设置 deleted_at 时间戳
func (r *TodoRepository) Delete(todoID, userID int) (bool, error) {
	result, err := r.db.Exec(
		`UPDATE todos SET deleted_at = EXTRACT(EPOCH FROM NOW())::BIGINT
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		todoID, userID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *TodoRepository) Complete(todoID, userID int) (models.Todo, error) {
	var todo models.Todo
	err := r.db.QueryRow(
		`UPDATE todos SET completed = TRUE, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		 RETURNING id, user_id, title, completed, created_at, updated_at`,
		todoID, userID,
	).Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	return todo, err
}
