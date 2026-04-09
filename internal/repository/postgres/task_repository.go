package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskRepository struct {
	db DBExecutor
}

func NewTaskRepository(db DBExecutor) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, template_id, title, description, status, scheduled_for, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, template_id, title, description, status, scheduled_for, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrTaskNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, template_id, title, description, status, scheduled_for, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrTaskNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, template_id, title, description, status, scheduled_for, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) CreateByTemplate(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (template_id, title, description, status, scheduled_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (template_id, scheduled_for) WHERE template_id IS NOT NULL DO NOTHING
		RETURNING id, template_id, title, description, status, scheduled_for, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		task.TemplateID,
		task.Title,
		task.Description,
		task.Status,
		task.ScheduledFor,
		task.CreatedAt,
		task.UpdatedAt,
	)

	created, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			const query = `
				SELECT id, template_id, title, description, status, scheduled_for, created_at, updated_at
				FROM tasks
				WHERE template_id = $1 AND scheduled_for = $2
			`

			existingRow := r.db.QueryRow(ctx, query, task.TemplateID, task.ScheduledFor)
			return scanTask(existingRow)
		}
		return nil, err
	}
	return created, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.TemplateID,
		&task.Title,
		&task.Description,
		&status,
		&task.ScheduledFor,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
