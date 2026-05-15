package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/yourname/freelance-platform/job-service/proto/job"
)

type gatewayConfig struct {
	HTTPAddr    string
	JobGRPCAddr string
}

func loadConfig() gatewayConfig {
	return gatewayConfig{
		HTTPAddr:    getEnv("GATEWAY_HTTP_ADDR", ":8080"),
		JobGRPCAddr: getEnv("JOB_SERVICE_GRPC_ADDR", "127.0.0.1:50052"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type errorResponse struct {
	Error string `json:"error"`
}

type jobDTO struct {
	ID          string  `json:"id"`
	ClientID    string  `json:"client_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type applicationDTO struct {
	ID           string `json:"id"`
	JobID        string `json:"job_id"`
	FreelancerID string `json:"freelancer_id"`
	CoverLetter  string `json:"cover_letter"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

func jobToDTO(j *pb.Job) jobDTO {
	dto := jobDTO{
		ID:          j.GetId(),
		ClientID:    j.GetClientId(),
		Title:       j.GetTitle(),
		Description: j.GetDescription(),
		Budget:      j.GetBudget(),
		Status:      j.GetStatus().String(),
	}
	if ts := j.GetCreatedAt(); ts != nil {
		dto.CreatedAt = ts.AsTime().UTC().Format(time.RFC3339Nano)
	}
	if ts := j.GetUpdatedAt(); ts != nil {
		dto.UpdatedAt = ts.AsTime().UTC().Format(time.RFC3339Nano)
	}
	return dto
}

func appToDTO(a *pb.Application) applicationDTO {
	dto := applicationDTO{
		ID:           a.GetId(),
		JobID:        a.GetJobId(),
		FreelancerID: a.GetFreelancerId(),
		CoverLetter:  a.GetCoverLetter(),
		Status:       a.GetStatus().String(),
	}
	if ts := a.GetCreatedAt(); ts != nil {
		dto.CreatedAt = ts.AsTime().UTC().Format(time.RFC3339Nano)
	}
	return dto
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, statusCode int, msg string) {
	writeJSON(w, statusCode, errorResponse{Error: msg})
}

func httpStatusFromGRPC(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, err.Error()
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()
	case codes.NotFound:
		return http.StatusNotFound, st.Message()
	case codes.Unauthenticated:
		return http.StatusUnauthorized, st.Message()
	case codes.PermissionDenied:
		return http.StatusForbidden, st.Message()
	case codes.Unavailable:
		return http.StatusServiceUnavailable, st.Message()
	default:
		return http.StatusInternalServerError, st.Message()
	}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func parseIntQuery(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return n, nil
}

type server struct {
	job pb.JobServiceClient
}

func main() {
	cfg := loadConfig()

	conn, err := grpc.NewClient(cfg.JobGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial job-service gRPC (%s): %v", cfg.JobGRPCAddr, err)
	}
	defer conn.Close()

	s := &server{job: pb.NewJobServiceClient(conn)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /jobs", s.handleCreateJob)
	mux.HandleFunc("GET /jobs/{job_id}", s.handleGetJob)
	mux.HandleFunc("GET /jobs", s.handleListJobs)
	mux.HandleFunc("POST /jobs/{job_id}/apply", s.handleApplyToJob)
	mux.HandleFunc("POST /jobs/{job_id}/accept", s.handleAcceptFreelancer)

	log.Printf("API Gateway listening on %s (job grpc: %s)", cfg.HTTPAddr, cfg.JobGRPCAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createJobBody struct {
	ClientID    string  `json:"client_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
}

func (s *server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var body createJobBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.job.CreateJob(ctx, &pb.CreateJobRequest{
		ClientId:    body.ClientID,
		Title:       body.Title,
		Description: body.Description,
		Budget:      body.Budget,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusCreated, jobToDTO(resp.GetJob()))
}

func (s *server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job_id is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.job.GetJob(ctx, &pb.GetJobRequest{JobId: jobID})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusOK, jobToDTO(resp.GetJob()))
}

type listJobsResponse struct {
	Jobs  []jobDTO `json:"jobs"`
	Total int32    `json:"total"`
}

func (s *server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	page, err := parseIntQuery(r, "page", 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid page")
		return
	}
	pageSize, err := parseIntQuery(r, "page_size", 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid page_size")
		return
	}
	clientID := r.URL.Query().Get("client_id")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := s.job.ListJobs(ctx, &pb.ListJobsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		ClientId: clientID,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}

	out := listJobsResponse{Total: resp.GetTotal()}
	for _, j := range resp.GetJobs() {
		out.Jobs = append(out.Jobs, jobToDTO(j))
	}
	writeJSON(w, http.StatusOK, out)
}

type applyToJobBody struct {
	FreelancerID string `json:"freelancer_id"`
	CoverLetter  string `json:"cover_letter"`
}

func (s *server) handleApplyToJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job_id is required")
		return
	}

	var body applyToJobBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := s.job.ApplyToJob(ctx, &pb.ApplyToJobRequest{
		JobId:        jobID,
		FreelancerId: body.FreelancerID,
		CoverLetter:  body.CoverLetter,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusCreated, appToDTO(resp.GetApplication()))
}

type acceptFreelancerBody struct {
	ApplicationID string `json:"application_id"`
}

type acceptFreelancerResponse struct {
	Job         jobDTO         `json:"job"`
	Application applicationDTO `json:"application"`
}

func (s *server) handleAcceptFreelancer(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job_id is required")
		return
	}

	var body acceptFreelancerBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.ApplicationID == "" {
		writeError(w, http.StatusBadRequest, "application_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := s.job.AcceptFreelancer(ctx, &pb.AcceptFreelancerRequest{
		JobId:         jobID,
		ApplicationId: body.ApplicationID,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}

	out := acceptFreelancerResponse{
		Job:         jobToDTO(resp.GetJob()),
		Application: appToDTO(resp.GetApplication()),
	}
	writeJSON(w, http.StatusOK, out)
}
