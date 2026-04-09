package postgres

import (
	"context"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{
		pool: pool,
	}
}

type txStore struct {
	templateRepo *TemplateRepository
	taskRepo     *TaskRepository
}

func (tx *txStore) CreateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	return tx.templateRepo.CreateTemplate(ctx, template)
}

func (tx *txStore) SetSpecificDates(ctx context.Context, templateID int64, specificDates []time.Time) ([]time.Time, error) {
	return tx.templateRepo.SetSpecificDates(ctx, templateID, specificDates)
}

func (tx *txStore) UpdateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	return tx.templateRepo.UpdateTemplate(ctx, template)
}

func (tx *txStore) UpdateSpecificDates(ctx context.Context, templateID int64, specificDates []time.Time) ([]time.Time, error) {
	return tx.templateRepo.UpdateSpecificDates(ctx, templateID, specificDates)
}

func (tx *txStore) CreateByTemplate(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	return tx.taskRepo.CreateByTemplate(ctx, task)
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(store ports.TxRepositories) error) (err error) {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit tx: %w", commitErr)
		}
	}()

	templateRepo := NewTemplateRepository(tx)
	taskRepo := NewTaskRepository(tx)

	store := &txStore{
		templateRepo: templateRepo,
		taskRepo:     taskRepo,
	}

	err = fn(store)
	if err != nil {
		return err
	}

	return nil
}
