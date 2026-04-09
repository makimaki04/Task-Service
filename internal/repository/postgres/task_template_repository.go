package postgres

import (
	"context"
	"errors"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/jackc/pgx/v5"
)

type TemplateRepository struct {
	db DBExecutor
}

func NewTemplateRepository(db DBExecutor) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) CreateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	const query = `
		INSERT INTO task_templates (title, description, recurrence_type, every_n_days, day_of_month, month_parity, start_date, end_date, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, title, description, recurrence_type, every_n_days, day_of_month, month_parity, start_date, end_date, active, created_at, updated_at
	`

	everyD, dOfMonth, parity := mapRecurrenceToDBFields(template.Recurrence)

	row := r.db.QueryRow(
		ctx,
		query,
		template.Title,
		template.Description,
		template.Recurrence.Type,
		everyD,
		dOfMonth,
		parity,
		template.StartDate,
		template.EndDate,
		template.Active,
		template.CreatedAt,
		template.UpdatedAt,
	)

	created, err := scanTemplate(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *TemplateRepository) SetSpecificDates(ctx context.Context, templateID int64, specificDates []time.Time) ([]time.Time, error) {
	dates := make([]time.Time, 0, len(specificDates))

	for _, sd := range specificDates {
		date, err := r.setSpecificDate(ctx, templateID, sd)
		if err != nil {
			return []time.Time{}, err
		}

		dates = append(dates, date)
	}

	return dates, nil
}

func (r *TemplateRepository) setSpecificDate(ctx context.Context, templateID int64, specificDate time.Time) (time.Time, error) {
	const query = `
		INSERT INTO task_template_dates (template_id, scheduled_date)
		VALUES ($1, $2)
		RETURNING scheduled_date
	`

	var date time.Time
	err := r.db.QueryRow(
		ctx,
		query,
		templateID,
		specificDate,
	).Scan(&date)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

func (r *TemplateRepository) UpdateSpecificDates(ctx context.Context, templateID int64, specificDates []time.Time) ([]time.Time, error) {
	const query = `
		DELETE FROM task_template_dates
		WHERE template_id = $1
	`
	_, err := r.db.Exec(
		ctx,
		query,
		templateID,
	)
	if err != nil {
		return []time.Time{}, err
	}

	dates := make([]time.Time, 0, len(specificDates))
	for _, sd := range specificDates {
		date, err := r.setSpecificDate(ctx, templateID, sd)
		if err != nil {
			return []time.Time{}, err
		}

		dates = append(dates, date)
	}

	return dates, nil
}

func (r *TemplateRepository) UpdateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	const query = `
		UPDATE task_templates 
		SET
		title=$2, 
		description=$3, 
		recurrence_type=$4, 
		every_n_days=$5, 
		day_of_month=$6, 
		month_parity=$7, 
		start_date=$8, 
		end_date=$9, 
		updated_at=$10
		WHERE id = $1
		RETURNING id, title, description, recurrence_type, every_n_days, day_of_month, month_parity, start_date, end_date, active, created_at, updated_at
	`

	everyD, dOfMonth, parity := mapRecurrenceToDBFields(template.Recurrence)

	row := r.db.QueryRow(
		ctx,
		query,
		template.ID,
		template.Title,
		template.Description,
		template.Recurrence.Type,
		everyD,
		dOfMonth,
		parity,
		template.StartDate,
		template.EndDate,
		template.UpdatedAt,
	)

	created, err := scanTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return created, nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, recurrence_type, every_n_days, day_of_month, month_parity, start_date, end_date, active, created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	found, err := scanTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	if found.Recurrence.Type == taskdomain.RecurrenceSpecificDates {
		dates, err := r.getSpecificDates(ctx, found.ID)
		if err != nil {
			return nil, err
		}
		found.Recurrence.SpecificDates = dates
	}

	return found, nil
}

func (r *TemplateRepository) List(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, recurrence_type, every_n_days, day_of_month, month_parity, start_date, end_date, active, created_at, updated_at
		FROM task_templates
		ORDER BY id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]taskdomain.TaskTemplate, 0)
	for rows.Next() {
		template, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}

		if template.Recurrence.Type == taskdomain.RecurrenceSpecificDates {
			dates, err := r.getSpecificDates(ctx, template.ID)
			if err != nil {
				return nil, err
			}
			template.Recurrence.SpecificDates = dates
		}

		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id int64) error {
	const query = `
		UPDATE task_templates
		SET active = FALSE
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *TemplateRepository) getSpecificDates(ctx context.Context, templateID int64) ([]time.Time, error) {
	const query = `
		SELECT scheduled_date
		FROM task_template_dates
		WHERE template_id = $1
		ORDER BY scheduled_date ASC
	`

	rows, err := r.db.Query(ctx, query, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dates, nil
}

func mapRecurrenceToDBFields(recurrence taskdomain.Recurrence) (*int, *int, *string) {
	switch recurrence.Type {
	case taskdomain.RecurrenceDaily:
		return &recurrence.EveryNDays, nil, nil
	case taskdomain.RecurrenceMonthly:
		return nil, &recurrence.DayOfMonth, nil
	case taskdomain.RecurrenceMonthParity:
		return nil, nil, (*string)(&recurrence.MonthParity)
	case taskdomain.RecurrenceSpecificDates:
		return nil, nil, nil
	default:
		return nil, nil, nil
	}
}

type templateScanner interface {
	Scan(dest ...any) error
}

func scanTemplate(scanner templateScanner) (*taskdomain.TaskTemplate, error) {
	var template taskdomain.TaskTemplate

	if err := scanner.Scan(
		&template.ID,
		&template.Title,
		&template.Description,
		&template.Recurrence.Type,
		&template.Recurrence.EveryNDays,
		&template.Recurrence.DayOfMonth,
		&template.Recurrence.MonthParity,
		&template.StartDate,
		&template.EndDate,
		&template.Active,
		&template.CreatedAt,
		&template.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &template, nil
}
