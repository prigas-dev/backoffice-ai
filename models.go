package main

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatus represents the status of a task in the Kanban board
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

// Task represents a task in the Kanban board
type Task struct {
	gorm.Model
	Title       string
	Description string
	Status      TaskStatus
	Priority    int
	DueDate     *time.Time
}

// User represents a user who can be assigned to tasks
type User struct {
	gorm.Model
	Name  string
	Email string
	Tasks []Task `gorm:"many2many:user_tasks;"`
}

type Page struct {
	gorm.Model
	Title string
	Path  string
}

type PageComponent struct {
	gorm.Model

	Page   Page
	PageID uint

	Component   Component
	ComponentID uint

	Query string
}

type Component struct {
	gorm.Model

	HTML string
}
