package usecase_test

import (
	"errors"
	"testing"

	"github.com/yourname/freelance-platform/job-service/internal/domain"
	"github.com/yourname/freelance-platform/job-service/internal/usecase"
)

// ─────────────────────────────────────────────
// Mocks
// ─────────────────────────────────────────────

type mockJobRepo struct {
	jobs map[string]*domain.Job
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{jobs: make(map[string]*domain.Job)}
}

func (m *mockJobRepo) Create(job *domain.Job) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepo) GetByID(id string) (*domain.Job, error) {
	if j, ok := m.jobs[id]; ok {
		return j, nil
	}
	return nil, errors.New("job not found")
}

func (m *mockJobRepo) List(page, pageSize int, clientID string) ([]*domain.Job, int, error) {
	var result []*domain.Job
	for _, j := range m.jobs {
		if clientID == "" || j.ClientID == clientID {
			result = append(result, j)
		}
	}
	return result, len(result), nil
}

func (m *mockJobRepo) UpdateStatus(id string, status domain.JobStatus) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = status
		return nil
	}
	return errors.New("job not found")
}

// ─────────────────────────────────────────────

type mockAppRepo struct {
	apps map[string]*domain.Application
}

func newMockAppRepo() *mockAppRepo {
	return &mockAppRepo{apps: make(map[string]*domain.Application)}
}

func (m *mockAppRepo) Create(app *domain.Application) error {
	m.apps[app.ID] = app
	return nil
}

func (m *mockAppRepo) GetByID(id string) (*domain.Application, error) {
	if a, ok := m.apps[id]; ok {
		return a, nil
	}
	return nil, errors.New("application not found")
}

func (m *mockAppRepo) ListByJob(jobID string) ([]*domain.Application, error) {
	var result []*domain.Application
	for _, a := range m.apps {
		if a.JobID == jobID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAppRepo) UpdateStatus(id string, status domain.ApplicationStatus) error {
	if a, ok := m.apps[id]; ok {
		a.Status = status
		return nil
	}
	return errors.New("application not found")
}

func (m *mockAppRepo) AcceptWithTx(applicationID, jobID string) error {
	if a, ok := m.apps[applicationID]; ok {
		a.Status = domain.ApplicationStatusAccepted
		return nil
	}
	return errors.New("application not found")
}

// ─────────────────────────────────────────────

type mockPublisher struct{ published int }

func (m *mockPublisher) PublishJobAccepted(_, _, _ string) error {
	m.published++
	return nil
}

type mockEmail struct{ sent int }

func (m *mockEmail) SendApplicationReceived(_, _, _ string) error { m.sent++; return nil }
func (m *mockEmail) SendFreelancerAccepted(_, _ string) error     { m.sent++; return nil }

// ─────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────

func newUC() (*usecase.JobUseCase, *mockJobRepo, *mockAppRepo, *mockPublisher, *mockEmail) {
	jr := newMockJobRepo()
	ar := newMockAppRepo()
	pub := &mockPublisher{}
	em := &mockEmail{}
	uc := usecase.NewJobUseCase(jr, ar, pub, em)
	return uc, jr, ar, pub, em
}

func TestCreateJob_Success(t *testing.T) {
	uc, _, _, _, _ := newUC()
	job, err := uc.CreateJob("client-1", "Build an API", "Go REST API", 1500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Error("expected job ID to be set")
	}
	if job.Status != domain.JobStatusOpen {
		t.Errorf("expected status open, got %s", job.Status)
	}
}

func TestCreateJob_MissingTitle(t *testing.T) {
	uc, _, _, _, _ := newUC()
	_, err := uc.CreateJob("client-1", "", "desc", 500)
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestCreateJob_NegativeBudget(t *testing.T) {
	uc, _, _, _, _ := newUC()
	_, err := uc.CreateJob("client-1", "Title", "desc", -100)
	if err == nil {
		t.Error("expected error for negative budget")
	}
}

func TestGetJob_NotFound(t *testing.T) {
	uc, _, _, _, _ := newUC()
	_, err := uc.GetJob("nonexistent-id")
	if err == nil {
		t.Error("expected not found error")
	}
}

func TestApplyToJob_Success(t *testing.T) {
	uc, _, _, _, _ := newUC()
	job, _ := uc.CreateJob("client-1", "Build API", "desc", 1000)

	app, err := uc.ApplyToJob(job.ID, "freelancer-1", "I can do this")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.JobID != job.ID {
		t.Errorf("expected job ID %s, got %s", job.ID, app.JobID)
	}
	if app.Status != domain.ApplicationStatusPending {
		t.Errorf("expected pending status, got %s", app.Status)
	}
}

func TestApplyToJob_ClosedJob(t *testing.T) {
	uc, jr, _, _, _ := newUC()
	job, _ := uc.CreateJob("client-1", "Build API", "desc", 1000)
	jr.UpdateStatus(job.ID, domain.JobStatusClosed)

	_, err := uc.ApplyToJob(job.ID, "freelancer-1", "cover letter")
	if err == nil {
		t.Error("expected error when applying to closed job")
	}
}

func TestAcceptFreelancer_Success(t *testing.T) {
	uc, _, _, pub, _ := newUC()
	job, _ := uc.CreateJob("client-1", "Build API", "desc", 1000)
	app, _ := uc.ApplyToJob(job.ID, "freelancer-1", "cover letter")

	updatedJob, updatedApp, err := uc.AcceptFreelancer(job.ID, app.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedJob.Status != domain.JobStatusInProgress {
		t.Errorf("expected in_progress, got %s", updatedJob.Status)
	}
	if updatedApp.Status != domain.ApplicationStatusAccepted {
		t.Errorf("expected accepted, got %s", updatedApp.Status)
	}
	if pub.published != 1 {
		t.Errorf("expected 1 event published, got %d", pub.published)
	}
}

func TestAcceptFreelancer_WrongJob(t *testing.T) {
	uc, _, _, _, _ := newUC()
	job1, _ := uc.CreateJob("client-1", "Job 1", "desc", 1000)
	job2, _ := uc.CreateJob("client-1", "Job 2", "desc", 2000)
	app, _ := uc.ApplyToJob(job1.ID, "freelancer-1", "cover letter")

	_, _, err := uc.AcceptFreelancer(job2.ID, app.ID)
	if err == nil {
		t.Error("expected error when application belongs to different job")
	}
}
