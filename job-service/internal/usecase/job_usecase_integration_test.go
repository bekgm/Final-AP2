package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/bekgm/Final-AP2/job-service/internal/domain"
	"github.com/bekgm/Final-AP2/job-service/internal/email"
	"github.com/bekgm/Final-AP2/job-service/internal/messaging"
	pgRepo "github.com/bekgm/Final-AP2/job-service/internal/repository/postgres"
	"github.com/bekgm/Final-AP2/job-service/internal/usecase"
)

// Integration test setup - requires running PostgreSQL
// Run with: docker compose up postgres-jobs -d
// Then: go test -tags=integration ./...

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/jobdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Test database not available: %v", err)
	}

	return db
}

func cleanupTestData(t *testing.T, db *sql.DB) {
	_, _ = db.Exec("DELETE FROM applications WHERE job_id IN (SELECT id FROM jobs WHERE client_id LIKE 'test-%')")
	_, _ = db.Exec("DELETE FROM jobs WHERE client_id LIKE 'test-%'")
}

func TestIntegrationCreateJob(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	jobRepo := pgRepo.NewJobRepository(db)
	appRepo := pgRepo.NewApplicationRepository(db)
	pub := &messaging.NoopPublisher{}
	emailer := &email.NoopSender{}
	uc := usecase.NewJobUseCase(jobRepo, appRepo, pub, emailer)

	clientID := fmt.Sprintf("test-client-%d", time.Now().UnixNano())
	
	job, err := uc.CreateJob(clientID, "Integration Test Job", "Test description", 5000)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if job.ID == "" {
		t.Error("Job ID should not be empty")
	}

	// Verify job was saved to DB
	found, err := uc.GetJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if found.Title != "Integration Test Job" {
		t.Errorf("Expected title 'Integration Test Job', got '%s'", found.Title)
	}
}

func TestIntegrationJobLifecycle(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	jobRepo := pgRepo.NewJobRepository(db)
	appRepo := pgRepo.NewApplicationRepository(db)
	pub := &messaging.NoopPublisher{}
	emailer := &email.NoopSender{}
	uc := usecase.NewJobUseCase(jobRepo, appRepo, pub, emailer)

	clientID := fmt.Sprintf("test-client-%d", time.Now().UnixNano())
	freelancerID := fmt.Sprintf("test-freelancer-%d", time.Now().UnixNano())

	// 1. Create job
	job, err := uc.CreateJob(clientID, "Full Lifecycle Test", "Complete workflow", 10000)
	if err != nil {
		t.Fatalf("Create job failed: %v", err)
	}

	// 2. Apply to job
	app, err := uc.ApplyToJob(job.ID, freelancerID, "I can do this job")
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if app.Status != domain.ApplicationStatusPending {
		t.Errorf("Expected pending status, got %s", app.Status)
	}

	// 3. List applications
	apps, err := uc.ListApplications(job.ID)
	if err != nil {
		t.Fatalf("List applications failed: %v", err)
	}

	if len(apps) != 1 {
		t.Errorf("Expected 1 application, got %d", len(apps))
	}

	// 4. Accept freelancer
	updatedJob, updatedApp, err := uc.AcceptFreelancer(job.ID, app.ID)
	if err != nil {
		t.Fatalf("Accept freelancer failed: %v", err)
	}

	if updatedJob.Status != domain.JobStatusInProgress {
		t.Errorf("Expected in_progress, got %s", updatedJob.Status)
	}

	if updatedApp.Status != domain.ApplicationStatusAccepted {
		t.Errorf("Expected accepted, got %s", updatedApp.Status)
	}

	// 5. Complete job
	completedJob, err := uc.CompleteJob(job.ID)
	if err != nil {
		t.Fatalf("Complete job failed: %v", err)
	}

	if completedJob.Status != domain.JobStatusClosed {
		t.Errorf("Expected closed, got %s", completedJob.Status)
	}

	t.Log("Full job lifecycle test passed!")
}

func TestIntegrationListJobs(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	jobRepo := pgRepo.NewJobRepository(db)
	appRepo := pgRepo.NewApplicationRepository(db)
	pub := &messaging.NoopPublisher{}
	emailer := &email.NoopSender{}
	uc := usecase.NewJobUseCase(jobRepo, appRepo, pub, emailer)

	clientID := fmt.Sprintf("test-client-%d", time.Now().UnixNano())

	// Create multiple jobs
	for i := 0; i < 5; i++ {
		_, err := uc.CreateJob(clientID, fmt.Sprintf("Job %d", i), "Description", 1000)
		if err != nil {
			t.Fatalf("Failed to create job %d: %v", i, err)
		}
	}

	// List jobs
	jobs, total, err := uc.ListJobs(1, 10, clientID)
	if err != nil {
		t.Fatalf("List jobs failed: %v", err)
	}

	if len(jobs) != 5 {
		t.Errorf("Expected 5 jobs, got %d", len(jobs))
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
}
