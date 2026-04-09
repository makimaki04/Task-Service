package task_template

import (
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestNormalizeAndValidateInput_SpecificDates(t *testing.T) {
	input := TemplateInput{
		Title:       "  Обзвон пациентов  ",
		Description: "  Утренний обход  ",
		Recurrence: taskdomain.Recurrence{
			Type: taskdomain.RecurrenceSpecificDates,
			SpecificDates: []time.Time{
				time.Date(2026, 4, 12, 14, 30, 0, 0, time.FixedZone("UTC+5", 5*60*60)),
				time.Date(2026, 4, 10, 11, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 12, 8, 0, 0, 0, time.UTC),
			},
		},
		StartDate: time.Date(2026, 4, 9, 13, 0, 0, 0, time.FixedZone("UTC+5", 5*60*60)),
	}

	got, err := normalizeAndValidateInput(input)
	if err != nil {
		t.Fatalf("normalizeAndValidateInput() error = %v", err)
	}

	if got.Title != "Обзвон пациентов" {
		t.Fatalf("expected trimmed title, got %q", got.Title)
	}

	if got.Description != "Утренний обход" {
		t.Fatalf("expected trimmed description, got %q", got.Description)
	}

	expectedStart := time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)
	if !got.StartDate.Equal(expectedStart) {
		t.Fatalf("expected normalized start date %v, got %v", expectedStart, got.StartDate)
	}

	if len(got.Recurrence.SpecificDates) != 2 {
		t.Fatalf("expected deduplicated specific dates length 2, got %d", len(got.Recurrence.SpecificDates))
	}

	expectedDates := []time.Time{
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
	}

	for i := range expectedDates {
		if !got.Recurrence.SpecificDates[i].Equal(expectedDates[i]) {
			t.Fatalf("expected specific date %v at index %d, got %v", expectedDates[i], i, got.Recurrence.SpecificDates[i])
		}
	}
}

func TestValidateInput_InvalidMonthlyDay(t *testing.T) {
	input := TemplateInput{
		Title:     "Отчет",
		StartDate: time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
		Recurrence: taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceMonthly,
			DayOfMonth: 31,
		},
	}

	_, err := normalizeAndValidateInput(input)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestBuildScheduledDates_DailyEveryTwoDays(t *testing.T) {
	now := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	template := taskdomain.TaskTemplate{
		Recurrence: taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceDaily,
			EveryNDays: 2,
		},
		StartDate: time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
	}

	got := buildScheduledDates(3, template, now)
	want := []time.Time{
		time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC),
	}

	assertDatesEqual(t, got, want)
}

func TestBuildScheduledDates_MonthlySkipsMissingDay(t *testing.T) {
	now := time.Date(2026, 1, 29, 10, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	template := taskdomain.TaskTemplate{
		Recurrence: taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceMonthly,
			DayOfMonth: 30,
		},
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   &endDate,
	}

	got := buildScheduledDates(3, template, now)
	want := []time.Time{
		time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	}

	assertDatesEqual(t, got, want)
}

func TestBuildScheduledDates_MonthParityEven(t *testing.T) {
	now := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	template := taskdomain.TaskTemplate{
		Recurrence: taskdomain.Recurrence{
			Type:        taskdomain.RecurrenceMonthParity,
			MonthParity: taskdomain.MonthParityEven,
		},
		StartDate: time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
	}

	got := buildScheduledDates(3, template, now)
	want := []time.Time{
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC),
	}

	assertDatesEqual(t, got, want)
}

func TestBuildScheduledDates_SpecificDatesWithinHorizon(t *testing.T) {
	now := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	template := taskdomain.TaskTemplate{
		Recurrence: taskdomain.Recurrence{
			Type: taskdomain.RecurrenceSpecificDates,
			SpecificDates: []time.Time{
				time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC),
			},
		},
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   &endDate,
	}

	got := buildScheduledDates(10, template, now)
	want := []time.Time{
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
	}

	assertDatesEqual(t, got, want)
}

func assertDatesEqual(t *testing.T, got, want []time.Time) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d dates, got %d", len(want), len(got))
	}

	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Fatalf("expected date %v at index %d, got %v", want[i], i, got[i])
		}
	}
}
