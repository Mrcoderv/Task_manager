package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"todo-app/models"
)

// ErrNotFound is returned when no todo matches the given ID.
var ErrNotFound = errors.New("todo not found")

type TodoRepository struct {
	pool *pgxpool.Pool
}

func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{pool: pool}
}

// CreateTodo inserts a row and returns it (RETURNING avoids a second query).
func (r *TodoRepository) CreateTodo(ctx context.Context, title string) (models.Todo, error) {
	const query = `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, completed, created_at`

	var t models.Todo
	err := r.pool.QueryRow(ctx, query, title).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
	if err != nil {
		return models.Todo{}, fmt.Errorf("create todo: %w", err)
	}
	return t, nil
}

// GetTodos returns all todos, oldest first.
func (r *TodoRepository) GetTodos(ctx context.Context) ([]models.Todo, error) {
	const query = `SELECT id, title, completed, created_at FROM todos ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	todos := []models.Todo{}
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todos: %w", err)
	}
	return todos, nil
}

// GetTodoByID returns one todo or ErrNotFound.
func (r *TodoRepository) GetTodoByID(ctx context.Context, id int64) (models.Todo, error) {
	const query = `SELECT id, title, completed, created_at FROM todos WHERE id = $1`

	var t models.Todo
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Todo{}, ErrNotFound
	}
	if err != nil {
		return models.Todo{}, fmt.Errorf("get todo %d: %w", id, err)
	}
	return t, nil
}

// UpdateTodo changes the title and returns the updated row.
func (r *TodoRepository) UpdateTodo(ctx context.Context, id int64, title string) (models.Todo, error) {
	const query = `
		UPDATE todos SET title = $1
		WHERE id = $2
		RETURNING id, title, completed, created_at`

	var t models.Todo
	err := r.pool.QueryRow(ctx, query, title, id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Todo{}, ErrNotFound
	}
	if err != nil {
		return models.Todo{}, fmt.Errorf("update todo %d: %w", id, err)
	}
	return t, nil
}

// ToggleTodo flips completed in SQL, so no read-then-write race.
func (r *TodoRepository) ToggleTodo(ctx context.Context, id int64) (models.Todo, error) {
	const query = `
		UPDATE todos SET completed = NOT completed
		WHERE id = $1
		RETURNING id, title, completed, created_at`

	var t models.Todo
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Todo{}, ErrNotFound
	}
	if err != nil {
		return models.Todo{}, fmt.Errorf("toggle todo %d: %w", id, err)
	}
	return t, nil
}

// DeleteTodo removes a row. Exec is used because we need no rows back.
func (r *TodoRepository) DeleteTodo(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete todo %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
