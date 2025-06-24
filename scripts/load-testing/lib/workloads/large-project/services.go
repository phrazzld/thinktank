package main

import (
	"fmt"
	"sync"
	"time"
)

// UserService provides user-related operations
type UserService struct {
	users  map[int]*User
	nextID int
	mutex  sync.RWMutex
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	service := &UserService{
		users:  make(map[int]*User),
		nextID: 1,
	}

	// Add some default users
	_ = service.CreateUser(&User{
		Name:   "John Doe",
		Email:  "john@example.com",
		Active: true,
	})
	_ = service.CreateUser(&User{
		Name:   "Jane Smith",
		Email:  "jane@example.com",
		Active: true,
	})

	return service
}

// CreateUser creates a new user
func (s *UserService) CreateUser(user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if email already exists
	for _, existingUser := range s.users {
		if existingUser.Email == user.Email {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	user.ID = s.nextID
	s.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true

	s.users[user.ID] = user
	return nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id int) *User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	userCopy := *user
	return &userCopy
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() []*User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	existingUser, exists := s.users[user.ID]
	if !exists {
		return fmt.Errorf("user with ID %d not found", user.ID)
	}

	// Preserve creation time
	user.CreatedAt = existingUser.CreatedAt
	user.UpdatedAt = time.Now()

	s.users[user.ID] = user
	return nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user with ID %d not found", id)
	}

	delete(s.users, id)
	return nil
}

// TaskService provides task-related operations
type TaskService struct {
	tasks  map[int]*Task
	nextID int
	mutex  sync.RWMutex
}

// NewTaskService creates a new task service
func NewTaskService() *TaskService {
	service := &TaskService{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}

	// Add some default tasks
	dueDate := time.Now().Add(7 * 24 * time.Hour)
	_ = service.CreateTask(&Task{
		Title:       "Setup development environment",
		Description: "Install and configure development tools",
		Status:      TaskStatusTodo,
		Priority:    PriorityHigh,
		AssigneeID:  1,
		DueDate:     &dueDate,
	})

	_ = service.CreateTask(&Task{
		Title:       "Write API documentation",
		Description: "Document all REST API endpoints",
		Status:      TaskStatusInProgress,
		Priority:    PriorityMedium,
		AssigneeID:  2,
	})

	return service
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(task *Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	task.ID = s.nextID
	s.nextID++
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	s.tasks[task.ID] = task
	return nil
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(id int) *Task {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	taskCopy := *task
	return &taskCopy
}

// GetAllTasks retrieves all tasks
func (s *TaskService) GetAllTasks() []*Task {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}

	return tasks
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(task *Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	existingTask, exists := s.tasks[task.ID]
	if !exists {
		return fmt.Errorf("task with ID %d not found", task.ID)
	}

	// Preserve creation time
	task.CreatedAt = existingTask.CreatedAt
	task.UpdatedAt = time.Now()

	s.tasks[task.ID] = task
	return nil
}

// DeleteTask deletes a task by ID
func (s *TaskService) DeleteTask(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task with ID %d not found", id)
	}

	delete(s.tasks, id)
	return nil
}

// GetTasksByUser retrieves all tasks assigned to a specific user
func (s *TaskService) GetTasksByUser(userID int) []*Task {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var userTasks []*Task
	for _, task := range s.tasks {
		if task.AssigneeID == userID {
			taskCopy := *task
			userTasks = append(userTasks, &taskCopy)
		}
	}

	return userTasks
}

// GetOverdueTasks retrieves all overdue tasks
func (s *TaskService) GetOverdueTasks() []*Task {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var overdueTasks []*Task
	for _, task := range s.tasks {
		if task.IsOverdue() {
			taskCopy := *task
			overdueTasks = append(overdueTasks, &taskCopy)
		}
	}

	return overdueTasks
}
