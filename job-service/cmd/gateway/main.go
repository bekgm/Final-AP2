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

	userpb "github.com/yourname/freelance-platform/job-service/proto/user"
	msgpb "github.com/yourname/freelance-platform/job-service/proto/messaging"
)

type gatewayConfig struct {
	HTTPAddr            string
	JobGRPCAddr         string
	UserGRPCAddr        string
	MessagingGRPCAddr   string
}

func loadConfig() gatewayConfig {
	return gatewayConfig{
		HTTPAddr:          getEnv("GATEWAY_HTTP_ADDR", ":8080"),
		JobGRPCAddr:       getEnv("JOB_SERVICE_GRPC_ADDR", "127.0.0.1:50052"),
		UserGRPCAddr:      getEnv("USER_SERVICE_GRPC_ADDR", "127.0.0.1:50051"),
		MessagingGRPCAddr: getEnv("MESSAGING_SERVICE_GRPC_ADDR", "127.0.0.1:50053"),
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
	job       pb.JobServiceClient
	user      userpb.UserServiceClient
	messaging msgpb.MessagingServiceClient
}

func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial gRPC (%s): %v", addr, err)
	}
	return conn
}

func main() {
	cfg := loadConfig()

	jobConn := mustDial(cfg.JobGRPCAddr)
	defer jobConn.Close()

	userConn := mustDial(cfg.UserGRPCAddr)
	defer userConn.Close()

	msgConn := mustDial(cfg.MessagingGRPCAddr)
	defer msgConn.Close()

	s := &server{
		job:       pb.NewJobServiceClient(jobConn),
		user:      userpb.NewUserServiceClient(userConn),
		messaging: msgpb.NewMessagingServiceClient(msgConn),
	}

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Job routes
	mux.HandleFunc("POST /jobs", s.handleCreateJob)
	mux.HandleFunc("GET /jobs/{job_id}", s.handleGetJob)
	mux.HandleFunc("GET /jobs", s.handleListJobs)
	mux.HandleFunc("POST /jobs/{job_id}/apply", s.handleApplyToJob)
	mux.HandleFunc("POST /jobs/{job_id}/accept", s.handleAcceptFreelancer)
	mux.HandleFunc("POST /jobs/{job_id}/complete", s.handleCompleteJob)

	// User routes
	mux.HandleFunc("POST /users/register", s.handleRegister)
	mux.HandleFunc("POST /users/login", s.handleLogin)
	mux.HandleFunc("GET /users/{user_id}", s.handleGetUser)
	mux.HandleFunc("PATCH /users/{user_id}", s.handleUpdateUser)

	// Messaging routes
	mux.HandleFunc("POST /api/messages", s.handleSendMessage)
	mux.HandleFunc("GET /api/messages", s.handleGetMessages)
	mux.HandleFunc("GET /api/dialogs", s.handleGetDialogs)

	log.Printf("API Gateway listening on %s | job=%s user=%s messaging=%s",
		cfg.HTTPAddr, cfg.JobGRPCAddr, cfg.UserGRPCAddr, cfg.MessagingGRPCAddr)
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

func (s *server) handleCompleteJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("job_id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := s.job.CompleteJob(ctx, &pb.CompleteJobRequest{
		JobId: jobID,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}

	writeJSON(w, http.StatusOK, jobToDTO(resp.GetJob()))
}

// ── User handlers ────────────────────────────────────────────────────────────

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body registerBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	roleVal := userpb.Role_ROLE_CLIENT
	if body.Role == "freelancer" {
		roleVal = userpb.Role_ROLE_FREELANCER
	}

	resp, err := s.user.Register(ctx, &userpb.RegisterRequest{
		Email:    body.Email,
		Password: body.Password,
		Name:     body.Name,
		Role:     roleVal,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.user.Login(ctx, &userpb.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.user.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetUser())
}

type updateUserBody struct {
	Name      *string  `json:"name,omitempty"`
	Bio       *string  `json:"bio,omitempty"`
	Skills    []string `json:"skills,omitempty"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
}

func (s *server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	var body updateUserBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req := &userpb.UpdateUserRequest{
		UserId: userID,
		Skills: body.Skills,
	}
	if body.Name != nil {
		req.Name = body.Name
	}
	if body.Bio != nil {
		req.Bio = body.Bio
	}
	if body.AvatarURL != nil {
		req.AvatarUrl = body.AvatarURL
	}

	resp, err := s.user.UpdateUser(ctx, req)
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetUser())
}

// ── Messaging handlers ───────────────────────────────────────────────────────

type sendMessageBody struct {
	ReceiverID string `json:"receiver_id"`
	ProjectID  string `json:"project_id"`
	Content    string `json:"content"`
}

func (s *server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	senderID := r.Header.Get("X-User-ID")
	if senderID == "" {
		writeError(w, http.StatusUnauthorized, "X-User-ID header is required")
		return
	}
	var body sendMessageBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.messaging.SendMessage(ctx, &msgpb.SendMessageRequest{
		SenderId:   senderID,
		ReceiverId: body.ReceiverID,
		ProjectId:  body.ProjectID,
		Content:    body.Content,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusCreated, resp.GetMessage())
}

func (s *server) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	userID1 := r.Header.Get("X-User-ID")
	if userID1 == "" {
		writeError(w, http.StatusUnauthorized, "X-User-ID header is required")
		return
	}
	userID2 := r.URL.Query().Get("user_id")
	if userID2 == "" {
		writeError(w, http.StatusBadRequest, "user_id query param is required")
		return
	}
	projectID := r.URL.Query().Get("project_id")
	limit, _ := parseIntQuery(r, "limit", 50)
	offset, _ := parseIntQuery(r, "offset", 0)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.messaging.GetMessages(ctx, &msgpb.GetMessagesRequest{
		UserId_1:  userID1,
		UserId_2:  userID2,
		ProjectId: projectID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetMessages())
}

func (s *server) handleGetDialogs(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "X-User-ID header is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.messaging.GetDialogs(ctx, &msgpb.GetDialogsRequest{
		UserId: userID,
	})
	if err != nil {
		code, msg := httpStatusFromGRPC(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetDialogs())
}
