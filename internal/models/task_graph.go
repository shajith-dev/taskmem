package models

type TaskGraph struct {
	TaskID    int64 `db:"task_id"   json:"task_id"`
	DependsOn int64 `db:"depends_on" json:"depends_on"`
}
