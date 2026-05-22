package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yourname/freelance-platform/job-service/internal/domain"
	"github.com/yourname/freelance-platform/job-service/internal/usecase"
	pb "github.com/yourname/freelance-platform/job-service/proto/job"
)

type JobHandler struct {
	pb.UnimplementedJobServiceServer
	uc *usecase.JobUseCase
}

func NewJobHandler(uc *usecase.JobUseCase) *JobHandler {
	return &JobHandler{uc: uc}
}

func (h *JobHandler) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	job, err := h.uc.CreateJob(req.ClientId, req.Title, req.Description, req.Budget)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return &pb.CreateJobResponse{Job: domainJobToProto(job)}, nil
}

func (h *JobHandler) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	job, err := h.uc.GetJob(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	return &pb.GetJobResponse{Job: domainJobToProto(job)}, nil
}

func (h *JobHandler) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	jobs, total, err := h.uc.ListJobs(int(req.Page), int(req.PageSize), req.ClientId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var protoJobs []*pb.Job
	for _, j := range jobs {
		protoJobs = append(protoJobs, domainJobToProto(j))
	}
	return &pb.ListJobsResponse{Jobs: protoJobs, Total: int32(total)}, nil
}

func (h *JobHandler) ApplyToJob(ctx context.Context, req *pb.ApplyToJobRequest) (*pb.ApplyToJobResponse, error) {
	app, err := h.uc.ApplyToJob(req.JobId, req.FreelancerId, req.CoverLetter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return &pb.ApplyToJobResponse{Application: domainAppToProto(app)}, nil
}

func (h *JobHandler) AcceptFreelancer(ctx context.Context, req *pb.AcceptFreelancerRequest) (*pb.AcceptFreelancerResponse, error) {
	job, app, err := h.uc.AcceptFreelancer(req.JobId, req.ApplicationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return &pb.AcceptFreelancerResponse{
		Job:         domainJobToProto(job),
		Application: domainAppToProto(app),
	}, nil
}

func (h *JobHandler) ListApplications(ctx context.Context, req *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	apps, err := h.uc.ListApplications(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	var protoApps []*pb.Application
	for _, app := range apps {
		protoApps = append(protoApps, domainAppToProto(app))
	}
	return &pb.ListApplicationsResponse{Applications: protoApps}, nil
}

func (h *JobHandler) CompleteJob(ctx context.Context, req *pb.CompleteJobRequest) (*pb.CompleteJobResponse, error) {
	job, err := h.uc.CompleteJob(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return &pb.CompleteJobResponse{
		Job: domainJobToProto(job),
	}, nil
}

func domainJobToProto(j *domain.Job) *pb.Job {
	var st pb.JobStatus
	switch j.Status {
	case domain.JobStatusOpen:
		st = pb.JobStatus_JOB_STATUS_OPEN
	case domain.JobStatusInProgress:
		st = pb.JobStatus_JOB_STATUS_IN_PROGRESS
	case domain.JobStatusClosed:
		st = pb.JobStatus_JOB_STATUS_CLOSED
	default:
		st = pb.JobStatus_JOB_STATUS_UNSPECIFIED
	}
	return &pb.Job{
		Id:          j.ID,
		ClientId:    j.ClientID,
		Title:       j.Title,
		Description: j.Description,
		Budget:      j.Budget,
		Status:      st,
		CreatedAt:   timestamppb.New(j.CreatedAt),
		UpdatedAt:   timestamppb.New(j.UpdatedAt),
	}
}

func domainAppToProto(a *domain.Application) *pb.Application {
	var st pb.ApplicationStatus
	switch a.Status {
	case domain.ApplicationStatusPending:
		st = pb.ApplicationStatus_APPLICATION_STATUS_PENDING
	case domain.ApplicationStatusAccepted:
		st = pb.ApplicationStatus_APPLICATION_STATUS_ACCEPTED
	case domain.ApplicationStatusRejected:
		st = pb.ApplicationStatus_APPLICATION_STATUS_REJECTED
	default:
		st = pb.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED
	}
	return &pb.Application{
		Id:           a.ID,
		JobId:        a.JobID,
		FreelancerId: a.FreelancerID,
		CoverLetter:  a.CoverLetter,
		Status:       st,
		CreatedAt:    timestamppb.New(a.CreatedAt),
	}
}
