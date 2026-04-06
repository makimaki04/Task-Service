DROP INDEX IF EXISTS idx_task_templates_active;
DROP INDEX IF EXISTS idx_tasks_scheduled_for;
DROP INDEX IF EXISTS idx_tasks_template_id_scheduled_for;

ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS fk_tasks_template_id,
DROP COLUMN IF EXISTS scheduled_for,
DROP COLUMN IF EXISTS template_id;

DROP TABLE IF EXISTS task_template_dates;
DROP TABLE IF EXISTS task_templates;
