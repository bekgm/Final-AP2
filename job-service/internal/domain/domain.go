package domain

import (
	"time"
)

type JobStatus string

const (
	JobStatusOpen       JobStatus = "open"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusClosed     JobStatus = "closed"
)

type ApplicationStatus string

const (
	ApplicationStatusPending  ApplicationStatus = "pending"
	ApplicationStatusAccepted ApplicationStatus = "accepted"
	ApplicationStatusRejected ApplicationStatus = "rejected"
)

type Job struct {
	ID          string
	ClientID    string
	Title       string
	Description string
	Budget      float64
	Status      JobStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Application struct {
	ID           string
	JobID        string
	FreelancerID string
	CoverLetter  string
	Status       ApplicationStatus
	CreatedAt    time.Time
}

// ─────────────────────────────────────────────
// Repository interfaces
// ─────────────────────────────────────────────

type JobRepository interface {
	Create(job *Job) error
	GetByID(id string) (*Job, error)
	List(page, pageSize int, clientID string) ([]*Job, int, error)
	UpdateStatus(id string, status JobStatus) error
}

type ApplicationRepository interface {
	Create(app *Application) error
	GetByID(id string) (*Application, error)
	ListByJob(jobID string) ([]*Application, error)
	UpdateStatus(id string, status ApplicationStatus) error
}

// ─────────────────────────────────────────────
// Event publisher interface
// ─────────────────────────────────────────────

type EventPublisher interface {
	PublishJobAccepted(jobID, freelancerID, clientID string) error
}

// ─────────────────────────────────────────────
// Email sender interface
// ─────────────────────────────────────────────

type EmailSender interface {
	SendApplicationReceived(toEmail, jobTitle, freelancerName string) error
	SendFreelancerAccepted(toEmail, jobTitle string) error
}
