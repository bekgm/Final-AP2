package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/yourname/freelance-platform/job-service/internal/domain"
)

// ─────────────────────────────────────────────
// Job Repository
// ─────────────────────────────────────────────

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *domain.Job) error {
	query := `
		INSERT INTO jobs (id, client_id, title, description, budget, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query,
		job.ID, job.ClientID, job.Title, job.Description,
		job.Budget, job.Status, job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *JobRepository) GetByID(id string) (*domain.Job, error) {
	query := `SELECT id, client_id, title, description, budget, status, created_at, updated_at FROM jobs WHERE id = $1`
	row := r.db.QueryRow(query, id)

	job := &domain.Job{}
	err := row.Scan(&job.ID, &job.ClientID, &job.Title, &job.Description,
		&job.Budget, &job.Status, &job.CreatedAt, &job.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	return job, err
}

func (r *JobRepository) List(page, pageSize int, clientID string) ([]*domain.Job, int, error) {
	offset := (page - 1) * pageSize

	var (
		rows  *sql.Rows
		err   error
		total int
	)

	if clientID != "" {
		err = r.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE client_id = $1`, clientID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.db.Query(`
			SELECT id, client_id, title, description, budget, status, created_at, updated_at
			FROM jobs WHERE client_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			clientID, pageSize, offset)
	} else {
		err = r.db.QueryRow(`SELECT COUNT(*) FROM jobs`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.db.Query(`
			SELECT id, client_id, title, description, budget, status, created_at, updated_at
			FROM jobs ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []*domain.Job
	for rows.Next() {
		job := &domain.Job{}
		if err := rows.Scan(&job.ID, &job.ClientID, &job.Title, &job.Description,
			&job.Budget, &job.Status, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, job)
	}
	return jobs, total, nil
}

func (r *JobRepository) UpdateStatus(id string, status domain.JobStatus) error {
	_, err := r.db.Exec(`UPDATE jobs SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

// ─────────────────────────────────────────────
// Application Repository
// ─────────────────────────────────────────────

type ApplicationRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(app *domain.Application) error {
	query := `
		INSERT INTO applications (id, job_id, freelancer_id, cover_letter, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query,
		app.ID, app.JobID, app.FreelancerID, app.CoverLetter, app.Status, app.CreatedAt,
	)
	return err
}

func (r *ApplicationRepository) GetByID(id string) (*domain.Application, error) {
	query := `SELECT id, job_id, freelancer_id, cover_letter, status, created_at FROM applications WHERE id = $1`
	row := r.db.QueryRow(query, id)

	app := &domain.Application{}
	err := row.Scan(&app.ID, &app.JobID, &app.FreelancerID, &app.CoverLetter, &app.Status, &app.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("application not found: %s", id)
	}
	return app, err
}

func (r *ApplicationRepository) ListByJob(jobID string) ([]*domain.Application, error) {
	rows, err := r.db.Query(`
		SELECT id, job_id, freelancer_id, cover_letter, status, created_at
		FROM applications WHERE job_id = $1 ORDER BY created_at DESC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*domain.Application
	for rows.Next() {
		app := &domain.Application{}
		if err := rows.Scan(&app.ID, &app.JobID, &app.FreelancerID, &app.CoverLetter, &app.Status, &app.CreatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// UpdateStatus wraps both application + job status changes in a single transaction.
// Called by AcceptFreelancer use case.
func (r *ApplicationRepository) UpdateStatus(id string, status domain.ApplicationStatus) error {
	_, err := r.db.Exec(`UPDATE applications SET status = $1 WHERE id = $2`, status, id)
	return err
}

// AcceptWithTx satisfies domain.ApplicationRepository and runs atomically.
func (r *ApplicationRepository) AcceptWithTx(applicationID, jobID string) error {
	return AcceptWithTransaction(r.db, applicationID, jobID)
}

// AcceptWithTransaction atomically accepts one application and sets the job to in_progress.
func AcceptWithTransaction(db *sql.DB, applicationID string, jobID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE applications SET status = 'accepted' WHERE id = $1`, applicationID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE jobs SET status = 'in_progress', updated_at = NOW() WHERE id = $1`, jobID); err != nil {
		return err
	}

	return tx.Commit()
}
