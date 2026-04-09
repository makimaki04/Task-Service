package task_template

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/ports"
)

var (
	defaultHorizonDays int = 30
)

type TemplateService struct {
	templateRepo TemplateRepository
	txManager ports.TransactionManager
	now       func() time.Time
}

func NewTemplateService(templateRepo TemplateRepository, txManager ports.TransactionManager) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		txManager: txManager,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *TemplateService) Create(ctx context.Context, input TemplateInput) (*taskdomain.TaskTemplate, []*taskdomain.Task, error) {
	normalized, err := normalizeAndValidateInput(input)
	if err != nil {
		return nil, []*taskdomain.Task{}, err
	}

	model := &taskdomain.TaskTemplate{
		Title:       normalized.Title,
		Description: normalized.Description,
		Recurrence:  normalized.Recurrence,
		StartDate:   normalized.StartDate,
		EndDate:     normalized.EndDate,
		Active:      true,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	var createdTemplate *taskdomain.TaskTemplate
	var createdTasks []*taskdomain.Task
	if err := s.txManager.RunInTx(ctx, func(store ports.TxRepositories) error {
		created, createErr := store.CreateTemplate(ctx, model)
		if createErr != nil {
			return createErr
		}

		createdTemplate = created

		if createdTemplate.Recurrence.Type == taskdomain.RecurrenceSpecificDates {
			dates, err := store.SetSpecificDates(ctx, createdTemplate.ID, model.Recurrence.SpecificDates)
			if err != nil {
				return err
			}

			createdTemplate.Recurrence.SpecificDates = dates
		}

		var tasks []taskdomain.Task
		generated := generateTasks(defaultHorizonDays, *createdTemplate, now)
		tasks = append(tasks, generated...)

		createdTasks = make([]*taskdomain.Task, 0, len(tasks))
		for _, t := range tasks {
			task, err := store.CreateByTemplate(ctx, &t)
			if err != nil {
				return err
			}

			createdTasks = append(createdTasks, task)
		}

		return nil
	}); err != nil {
		return nil, []*taskdomain.Task{}, err
	}

	return createdTemplate, createdTasks, nil
}

func (s *TemplateService) Update(ctx context.Context, id int64, input TemplateInput) (*taskdomain.TaskTemplate, []*taskdomain.Task, error) {
	normalized, err := normalizeAndValidateInput(input)
	if err != nil {
		return nil, []*taskdomain.Task{}, err
	}

	model := &taskdomain.TaskTemplate{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Recurrence:  normalized.Recurrence,
		Active:      true,
		StartDate:   normalized.StartDate,
		EndDate:     normalized.EndDate,
	}
	now := s.now()
	model.UpdatedAt = now

	var updatedTemplate *taskdomain.TaskTemplate
	var createdTasks []*taskdomain.Task
	if err := s.txManager.RunInTx(ctx, func(store ports.TxRepositories) error {
		updated, updateErr := store.UpdateTemplate(ctx, model)
		if updateErr != nil {
			return updateErr
		}

		updatedTemplate = updated

		dates, err := store.UpdateSpecificDates(ctx, updatedTemplate.ID, model.Recurrence.SpecificDates)
		if err != nil {
			return err
		}

		if updatedTemplate.Recurrence.Type == taskdomain.RecurrenceSpecificDates {
			updatedTemplate.Recurrence.SpecificDates = dates
		} else {
			updatedTemplate.Recurrence.SpecificDates = nil
		}

		var tasks []taskdomain.Task
		generated := generateTasks(defaultHorizonDays, *updatedTemplate, now)
		tasks = append(tasks, generated...)

		createdTasks = make([]*taskdomain.Task, 0, len(tasks))
		for _, t := range tasks {
			task, err := store.CreateByTemplate(ctx, &t)
			if err != nil {
				return err
			}

			createdTasks = append(createdTasks, task)
		}

		return nil
	}); err != nil {
		return nil, []*taskdomain.Task{}, err
	}

	return updatedTemplate, createdTasks, nil
}

func (s *TemplateService) GetByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	return s.templateRepo.GetByID(ctx, id)
}

func (s *TemplateService) List(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	return s.templateRepo.List(ctx)
}

func (s *TemplateService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	return s.templateRepo.Delete(ctx, id)
}

func generateTasks(taskCount int, template taskdomain.TaskTemplate, now time.Time) []taskdomain.Task {
	tasks := make([]taskdomain.Task, 0, taskCount)
	scheduledDates := buildScheduledDates(taskCount, template, now)

	for _, scheduledFor := range scheduledDates {
		task := getTaskInstanceByTemplate(template, scheduledFor, now)
		tasks = append(tasks, task)
	}

	return tasks
}

func buildScheduledDates(limit int, template taskdomain.TaskTemplate, now time.Time) []time.Time {
	if limit <= 0 {
		return nil
	}

	today := normalizeDate(now)
	start := template.StartDate
	if start.Before(today) {
		start = today
	}
	var generationEnd *time.Time
	if template.EndDate == nil {
		endDate := today.AddDate(0, 0, defaultHorizonDays)
		generationEnd = &endDate
	} else {
		generationEnd = template.EndDate
	}

	var dates []time.Time

	switch template.Recurrence.Type {
	case taskdomain.RecurrenceDaily:
		for current := start; len(dates) < limit; current = current.AddDate(0, 0, template.Recurrence.EveryNDays) {
			if exceedsEndDate(current, generationEnd) {
				break
			}
			dates = append(dates, current)
		}

	case taskdomain.RecurrenceMonthly:
		year, month, _ := start.Date()
		currentMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

		for len(dates) < limit {
			daysInCurrentMonth := daysInMonth(currentMonth.Year(), currentMonth.Month())
			if template.Recurrence.DayOfMonth <= daysInCurrentMonth {
				candidate := time.Date(
					currentMonth.Year(),
					currentMonth.Month(),
					template.Recurrence.DayOfMonth,
					0,
					0,
					0,
					0,
					time.UTC,
				)
				if !candidate.Before(start) {
					if exceedsEndDate(candidate, generationEnd) {
						break
					}
					dates = append(dates, candidate)
				}
			}

			nextMonth := currentMonth.AddDate(0, 1, 0)
			firstDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			if exceedsEndDate(firstDayNextMonth, generationEnd) {
				break
			}
			currentMonth = firstDayNextMonth
		}

	case taskdomain.RecurrenceMonthParity:
		for current := start; len(dates) < limit; current = current.AddDate(0, 0, 1) {
			if exceedsEndDate(current, generationEnd) {
				break
			}
			day := current.Day()
			if template.Recurrence.MonthParity == taskdomain.MonthParityEven && day%2 == 0 {
				dates = append(dates, current)
			}
			if template.Recurrence.MonthParity == taskdomain.MonthParityOdd && day%2 != 0 {
				dates = append(dates, current)
			}
		}

	case taskdomain.RecurrenceSpecificDates:
		for _, scheduledFor := range template.Recurrence.SpecificDates {
			if scheduledFor.Before(start) {
				continue
			}
			if exceedsEndDate(scheduledFor, generationEnd) {
				continue
			}

			dates = append(dates, scheduledFor)
			if len(dates) == limit {
				break
			}
		}
	}

	return dates
}

func exceedsEndDate(date time.Time, endDate *time.Time) bool {
	if endDate == nil {
		return false
	}

	return date.After(*endDate)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func getTaskInstanceByTemplate(template taskdomain.TaskTemplate, scheduledFor time.Time, now time.Time) taskdomain.Task {
	task := taskdomain.Task{
		TemplateID:   &template.ID,
		Title:        template.Title,
		Description:  template.Description,
		Status:       taskdomain.StatusNew,
		ScheduledFor: &scheduledFor,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return task
}

func normalizeInput(input TemplateInput) TemplateInput {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	input.StartDate = normalizeDate(input.StartDate)
	if input.EndDate != nil {
		t := normalizeDate(*input.EndDate)
		input.EndDate = &t
	}

	input.Recurrence = normalizeRecurrence(input.Recurrence)

	return input
}

func normalizeRecurrence(r taskdomain.Recurrence) taskdomain.Recurrence {
	switch r.Type {
	case taskdomain.RecurrenceDaily:
		r.DayOfMonth = 0
		r.MonthParity = ""
		r.SpecificDates = nil
	case taskdomain.RecurrenceMonthly:
		r.EveryNDays = 0
		r.MonthParity = ""
		r.SpecificDates = nil
	case taskdomain.RecurrenceMonthParity:
		r.EveryNDays = 0
		r.DayOfMonth = 0
		r.SpecificDates = nil
	case taskdomain.RecurrenceSpecificDates:
		r.EveryNDays = 0
		r.DayOfMonth = 0
		r.MonthParity = ""
		r.SpecificDates = normalizeDates(r.SpecificDates)
	default:
	}

	return r
}

func normalizeDates(dates []time.Time) []time.Time {
	if len(dates) == 0 {
		return nil
	}

	res := make([]time.Time, 0, len(dates))
	seen := make(map[string]struct{}, len(dates))

	for _, d := range dates {
		d = normalizeDate(d)
		key := d.Format("2006-01-02")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		res = append(res, d)
	}

	slices.SortFunc(res, func(a, b time.Time) int {
		switch {
		case a.Before(b):
			return -1
		case a.After(b):
			return 1
		default:
			return 0
		}
	})

	return res
}

func normalizeDate(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}

	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func validateInput(input TemplateInput) error {
	if input.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}

	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return fmt.Errorf("%w: end_date must be greater than or equal to start_date", ErrInvalidInput)
	}

	if !input.Recurrence.Type.Valid() {
		return fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch input.Recurrence.Type {
	case taskdomain.RecurrenceDaily:
		if input.Recurrence.EveryNDays <= 0 {
			return fmt.Errorf("%w: every_n_days must be greater than 0", ErrInvalidInput)
		}

	case taskdomain.RecurrenceMonthly:
		if input.Recurrence.DayOfMonth < 1 || input.Recurrence.DayOfMonth > 30 {
			return fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}

	case taskdomain.RecurrenceMonthParity:
		if !input.Recurrence.MonthParity.Valid() {
			return fmt.Errorf("%w: invalid month_parity", ErrInvalidInput)
		}

	case taskdomain.RecurrenceSpecificDates:
		if len(input.Recurrence.SpecificDates) == 0 {
			return fmt.Errorf("%w: specific_dates is required", ErrInvalidInput)
		}
	}
	return nil
}

func normalizeAndValidateInput(input TemplateInput) (TemplateInput, error) {
	input = normalizeInput(input)

	if err := validateInput(input); err != nil {
		return TemplateInput{}, err
	}

	return input, nil
}
