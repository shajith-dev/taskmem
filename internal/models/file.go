package models

type File struct {
	ID       int64  `db:"id"        json:"id"`
	FilePath string `db:"file_path" json:"file_path"`
}

type TaskFile struct {
	TaskID int64 `db:"task_id" json:"task_id"`
	FileID int64 `db:"file_id" json:"file_id"`
}
