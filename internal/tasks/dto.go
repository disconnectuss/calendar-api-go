package tasks

type TaskStatus string

const (
	TaskStatusNeedsAction TaskStatus = "needsAction"
	TaskStatusCompleted   TaskStatus = "completed"
)

type CreateTaskRequest struct {
	Title  string     `json:"title" validate:"required"`
	Notes  string     `json:"notes,omitempty"`
	Due    string     `json:"due,omitempty" validate:"omitempty,rfc3339"`
	Status TaskStatus `json:"status,omitempty"`
}

type UpdateTaskRequest struct {
	Title     string     `json:"title,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
	Due       *string    `json:"due,omitempty" validate:"omitempty,rfc3339"`
	Status    TaskStatus `json:"status,omitempty"`
	Completed *bool      `json:"completed,omitempty"`
}

type CreateTaskListRequest struct {
	Title string `json:"title" validate:"required"`
}
