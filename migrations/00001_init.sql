-- +goose Up
-- +goose StatementBegin
CREATE TYPE task_status AS ENUM (
    'PENDING',
    'IN_PROGRESS',
    'COMPLETED',
    'PARTIALLY_COMPLETED'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE tasks (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent      BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    status      task_status NOT NULL DEFAULT 'PENDING',
    description TEXT NOT NULL,
    scratchpad  TEXT,
    model       TEXT NOT NULL DEFAULT 'inherit',
    use_subagent BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_tasks_parent ON tasks(parent);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE task_graph (
    task_id    BIGINT NOT NULL,
    depends_on BIGINT NOT NULL,

    PRIMARY KEY (task_id, depends_on),
    FOREIGN KEY (task_id)    REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on) REFERENCES tasks(id) ON DELETE CASCADE,
    CHECK (task_id <> depends_on)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE files (
    id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    file_path TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE task_files (
    task_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,

    PRIMARY KEY (task_id, file_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER tasks_set_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS tasks_set_updated_at ON tasks;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_at;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS task_files;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS task_graph;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS files;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS task_status;
-- +goose StatementEnd
