-- +goose Up
-- +goose StatementBegin
CREATE TABLE tasks (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    parent       INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'PENDING'
                 CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'PARTIALLY_COMPLETED')),
    description  TEXT NOT NULL,
    scratchpad   TEXT,
    model        TEXT NOT NULL DEFAULT 'inherit',
    use_subagent INTEGER NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at   DATETIME NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_tasks_parent ON tasks(parent);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE task_graph (
    task_id    INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,

    PRIMARY KEY (task_id, depends_on),
    CHECK (task_id <> depends_on)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE files (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE task_files (
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    file_id INTEGER NOT NULL REFERENCES files(id) ON DELETE CASCADE,

    PRIMARY KEY (task_id, file_id)
);
-- +goose StatementEnd

-- +goose Down
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
