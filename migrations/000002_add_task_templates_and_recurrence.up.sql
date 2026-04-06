CREATE TABLE IF NOT EXISTS task_templates (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    recurrence_type TEXT NOT NULL CHECK (recurrence_type IN ('daily', 'monthly', 'specific_dates', 'month_parity')),
    every_n_days INT NULL CHECK (every_n_days IS NULL OR every_n_days > 0),
    day_of_month INT NULL CHECK (day_of_month IS NULL OR day_of_month BETWEEN 1 AND 30),
    month_parity TEXT NULL CHECK (month_parity IS NULL OR month_parity IN ('even', 'odd')),
    start_date DATE NOT NULL,
    end_date DATE NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_date IS NULL OR end_date >= start_date),
    CHECK (
        (recurrence_type = 'daily' AND every_n_days IS NOT NULL AND day_of_month IS NULL AND month_parity IS NULL) OR
        (recurrence_type = 'monthly' AND every_n_days IS NULL AND day_of_month IS NOT NULL AND month_parity IS NULL) OR
        (recurrence_type = 'specific_dates' AND every_n_days IS NULL AND day_of_month IS NULL AND month_parity IS NULL) OR
        (recurrence_type = 'month_parity' AND every_n_days IS NULL AND day_of_month IS NULL AND month_parity IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS task_template_dates (
    template_id BIGINT NOT NULL REFERENCES task_templates(id) ON DELETE CASCADE,
    scheduled_date DATE NOT NULL,
    PRIMARY KEY (template_id, scheduled_date)
);

ALTER TABLE tasks
ADD COLUMN template_id BIGINT NULL,
ADD COLUMN scheduled_for DATE NULL,
ADD CONSTRAINT fk_tasks_template_id
    FOREIGN KEY (template_id) REFERENCES task_templates(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_template_id_scheduled_for
    ON tasks (template_id, scheduled_for)
    WHERE template_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for ON tasks (scheduled_for);
CREATE INDEX IF NOT EXISTS idx_task_templates_active ON task_templates (active);
