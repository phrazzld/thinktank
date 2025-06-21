package main

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Active    bool      `json:"active"`
}

// Task represents a task in the system
type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Priority    Priority   `json:"priority"`
	AssigneeID  int        `json:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Priority represents the priority level of a task
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Validate validates a User struct
func (u *User) Validate() error {
	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if u.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	// Simple email validation
	if !isValidEmail(u.Email) {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}
	return nil
}

// Validate validates a Task struct
func (t *Task) Validate() error {
	if t.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	if t.Status == "" {
		t.Status = TaskStatusTodo // Default status
	}
	if t.Priority == "" {
		t.Priority = PriorityMedium // Default priority
	}
	if !isValidTaskStatus(t.Status) {
		return &ValidationError{Field: "status", Message: "invalid task status"}
	}
	if !isValidPriority(t.Priority) {
		return &ValidationError{Field: "priority", Message: "invalid priority"}
	}
	return nil
}

// IsComplete returns true if the task is completed
func (t *Task) IsComplete() bool {
	return t.Status == TaskStatusDone
}

// IsOverdue returns true if the task is overdue
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	return time.Now().After(*t.DueDate) && !t.IsComplete()
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Helper functions
func isValidEmail(email string) bool {
	// Simple email validation - in production use proper regex or library
	return len(email) > 3 &&
		containsChar(email, '@') &&
		containsChar(email, '.')
}

func containsChar(s string, c rune) bool {
	for _, char := range s {
		if char == c {
			return true
		}
	}
	return false
}

func isValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

func isValidPriority(priority Priority) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}
