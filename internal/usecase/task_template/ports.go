package task_template

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TemplateRepository interface {
	GetByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error)
	List(ctx context.Context) ([]taskdomain.TaskTemplate, error)
	Delete(ctx context.Context, id int64) error
}

type Usecase interface {
	Create(ctx context.Context, input TemplateInput) (*taskdomain.TaskTemplate, []*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error)
	List(ctx context.Context) ([]taskdomain.TaskTemplate, error)
	Update(ctx context.Context, id int64, input TemplateInput) (*taskdomain.TaskTemplate, []*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
}

type TemplateInput struct {
	Title       string
	Description string
	Recurrence  taskdomain.Recurrence
	StartDate   time.Time
	EndDate     *time.Time
}
