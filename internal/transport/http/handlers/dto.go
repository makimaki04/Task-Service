package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID           int64              `json:"id"`
	TemplateID   *int64             `json:"template_id,omitempty"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Status       taskdomain.Status  `json:"status"`
	ScheduledFor *time.Time         `json:"scheduled_for,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type recurrenceDTO struct {
	Type          taskdomain.RecurrenceType `json:"type"`
	EveryNDays    int                       `json:"every_n_days,omitempty"`
	DayOfMonth    int                       `json:"day_of_month,omitempty"`
	MonthParity   taskdomain.MonthParity    `json:"month_parity,omitempty"`
	SpecificDates []time.Time               `json:"specific_dates,omitempty"`
}

type templateMutationDTO struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Recurrence  recurrenceDTO `json:"recurrence"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
}

type templateDTO struct {
	ID          int64         `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Recurrence  recurrenceDTO `json:"recurrence"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
	Active      bool          `json:"active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type templateCreateResponseDTO struct {
	Template templateDTO `json:"template"`
	Tasks    []taskDTO   `json:"tasks"`
}

type templateUpdateResponseDTO struct {
	Template templateDTO `json:"template"`
	Tasks    []taskDTO   `json:"tasks"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		TemplateID:   task.TemplateID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		ScheduledFor: task.ScheduledFor,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

func newTasksDTO(tasks []*taskdomain.Task) []taskDTO {
	response := make([]taskDTO, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, newTaskDTO(task))
	}

	return response
}

func newTemplateDTO(template *taskdomain.TaskTemplate) templateDTO {
	return templateDTO{
		ID:          template.ID,
		Title:       template.Title,
		Description: template.Description,
		Recurrence:  newRecurrenceDTO(template.Recurrence),
		StartDate:   template.StartDate,
		EndDate:     template.EndDate,
		Active:      template.Active,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}

func newRecurrenceDTO(recurrence taskdomain.Recurrence) recurrenceDTO {
	return recurrenceDTO{
		Type:          recurrence.Type,
		EveryNDays:    recurrence.EveryNDays,
		DayOfMonth:    recurrence.DayOfMonth,
		MonthParity:   recurrence.MonthParity,
		SpecificDates: recurrence.SpecificDates,
	}
}
