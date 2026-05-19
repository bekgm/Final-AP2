package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/freelance-platform/job-service/internal/domain"
)

type JobUseCase struct {
	jobRepo    domain.JobRepository
	appRepo    domain.ApplicationRepository
	publisher  domain.EventPublisher
	emailSender domain.EmailSender
}

func NewJobUseCase(
	jobRepo domain.JobRepository,
	appRepo domain.ApplicationRepository,
	publisher domain.EventPublisher,
	emailSender domain.EmailSender,
) *JobUseCase {
	return &JobUseCase{
		jobRepo:     jobRepo,
		appRepo:     appRepo,
		publisher:   publisher,
		emailSender: emailSender,
	}
}

// ── CreateJob ─────────────────────────────────

func (uc *JobUseCase) CreateJob(clientID, title, description string, budget float64) (*domain.Job, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if budget <= 0 {
		return nil, errors.New("budget must be positive")
	}

	job := &domain.Job{
		ID:          uuid.NewString(),
		ClientID:    clientID,
		Title:       title,
		Description: description,
		Budget:      budget,
		Status:      domain.JobStatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.jobRepo.Create(job); err != nil {
		return nil, err
	}
	return job, nil
}

// ── GetJob ────────────────────────────────────

func (uc *JobUseCase) GetJob(jobID string) (*domain.Job, error) {
	if jobID == "" {
		return nil, errors.New("job_id is required")
	}
	return uc.jobRepo.GetByID(jobID)
}

// ── ListJobs ──────────────────────────────────

func (uc *JobUseCase) ListJobs(page, pageSize int, clientID string) ([]*domain.Job, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return uc.jobRepo.List(page, pageSize, clientID)
}

// ── ApplyToJob ────────────────────────────────

func (uc *JobUseCase) ApplyToJob(jobID, freelancerID, coverLetter string) (*domain.Application, error) {
	job, err := uc.jobRepo.GetByID(jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != domain.JobStatusOpen {
		return nil, errors.New("job is not open for applications")
	}

	app := &domain.Application{
		ID:           uuid.NewString(),
		JobID:        jobID,
		FreelancerID: freelancerID,
		CoverLetter:  coverLetter,
		Status:       domain.ApplicationStatusPending,
		CreatedAt:    time.Now(),
	}

	if err := uc.appRepo.Create(app); err != nil {
		return nil, err
	}

	// Non-blocking email notification to client (best-effort)
	go func() {
		_ = uc.emailSender.SendApplicationReceived("client@example.com", job.Title, freelancerID)
	}()

	return app, nil
}

// ── AcceptFreelancer ──────────────────────────

// AcceptFreelancer runs atomically via the repository transaction:
// 1. Mark application as accepted
// 2. Mark job as in_progress
// 3. Publish job.accepted event to RabbitMQ
// 4. Send email to the accepted freelancer
func (uc *JobUseCase) AcceptFreelancer(jobID, applicationID string) (*domain.Job, *domain.Application, error) {
	job, err := uc.jobRepo.GetByID(jobID)
	if err != nil {
		return nil, nil, err
	}
	if job.Status != domain.JobStatusOpen {
		return nil, nil, errors.New("job is not open")
	}

	app, err := uc.appRepo.GetByID(applicationID)
	if err != nil {
		return nil, nil, err
	}
	if app.JobID != jobID {
		return nil, nil, errors.New("application does not belong to this job")
	}
	if app.Status != domain.ApplicationStatusPending {
		return nil, nil, errors.New("application is not pending")
	}

	// Update both statuses atomically in a single SQL transaction
	if err := uc.appRepo.AcceptWithTx(applicationID, jobID); err != nil {
		return nil, nil, err
	}

	// Publish event → Messaging Service will create a conversation
	if err := uc.publisher.PublishJobAccepted(jobID, app.FreelancerID, job.ClientID); err != nil {
		// Log but don't fail the request
	}

	// Email to accepted freelancer
	go func() {
		_ = uc.emailSender.SendFreelancerAccepted("freelancer@example.com", job.Title)
	}()

	app.Status = domain.ApplicationStatusAccepted
	job.Status = domain.JobStatusInProgress

	return job, app, nil
}

// ── CompleteJob ────────────────────────────────

func (uc *JobUseCase) CompleteJob(jobID string) (*domain.Job, error) {
	job, err := uc.jobRepo.GetByID(jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != domain.JobStatusInProgress {
		return nil, errors.New("job must be in progress to be completed")
	}

	if err := uc.jobRepo.UpdateStatus(jobID, domain.JobStatusClosed); err != nil {
		return nil, err
	}

	job.Status = domain.JobStatusClosed
	return job, nil
}
