package ports

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(repos TxRepositories) error) error
}

type TxRepositories interface {
	CreateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error)
	SetSpecificDates(ctx context.Context, templateID int64, specificDates []time.Time) ([]time.Time, error)
	UpdateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error)
	UpdateSpecificDates(ctx context.Context, templateID int64, specificDates []time.Time) ([]time.Time, error)
	CreateByTemplate(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
}
