package models

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusPending            TaskStatus = "PENDING"
	TaskStatusInProgress         TaskStatus = "IN_PROGRESS"
	TaskStatusCompleted          TaskStatus = "COMPLETED"
	TaskStatusPartiallyCompleted TaskStatus = "PARTIALLY_COMPLETED"
)

// Valid reports whether s is one of the four defined task statuses.
func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusPartiallyCompleted:
		return true
	}
	return false
}

type Task struct {
	ID          int64      `db:"id"          json:"id"`
	Parent      *int64     `db:"parent"       json:"parent,omitempty"`
	Status      TaskStatus `db:"status"       json:"status"`
	Description string     `db:"description"  json:"description"`
	Scratchpad  *string    `db:"scratchpad"   json:"scratchpad,omitempty"`
	Model       string     `db:"model"        json:"model"`
	UseSubagent bool       `db:"use_subagent" json:"use_subagent"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
}
