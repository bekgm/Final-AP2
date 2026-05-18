package service

import (
	"context"

	"github.com/bekgm/Final-AP2/internal/models"
	"github.com/bekgm/Final-AP2/internal/repository"
	pb "github.com/bekgm/Final-AP2/pkg/messaging"
)

type MessagingService struct {
	pb.UnimplementedMessagingServiceServer
	repo repository.MessagingRepository
}

func NewMessagingService(repo repository.MessagingRepository) *MessagingService {
	return &MessagingService{repo: repo}
}

func (s *MessagingService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	msg := &models.Message{
		SenderID:   req.SenderId,
		ReceiverID: req.ReceiverId,
		ProjectID:  req.ProjectId,
		Content:    req.Content,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	return &pb.SendMessageResponse{
		Message: &pb.Message{
			Id:         msg.ID,
			SenderId:   msg.SenderID,
			ReceiverId: msg.ReceiverID,
			ProjectId:  msg.ProjectID,
			Content:    msg.Content,
			Timestamp:  msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (s *MessagingService) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	offset := int(req.Offset)

	messages, err := s.repo.GetMessagesBetweenUsers(ctx, req.UserId_1, req.UserId_2, req.ProjectId, limit, offset)
	if err != nil {
		return nil, err
	}

	var pbMessages []*pb.Message
	for _, msg := range messages {
		pbMessages = append(pbMessages, &pb.Message{
			Id:         msg.ID,
			SenderId:   msg.SenderID,
			ReceiverId: msg.ReceiverID,
			ProjectId:  msg.ProjectID,
			Content:    msg.Content,
			Timestamp:  msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetMessagesResponse{Messages: pbMessages}, nil
}

func (s *MessagingService) GetDialogs(ctx context.Context, req *pb.GetDialogsRequest) (*pb.GetDialogsResponse, error) {
	dialogs, err := s.repo.GetDialogsForUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var pbDialogs []*pb.Dialog
	for _, d := range dialogs {
		pbDialogs = append(pbDialogs, &pb.Dialog{
			OtherUserId: d.OtherUserID,
			ProjectId:   d.ProjectID,
			LastMessage: &pb.Message{
				Id:         d.LastMessage.ID,
				SenderId:   d.LastMessage.SenderID,
				ReceiverId: d.LastMessage.ReceiverID,
				ProjectId:  d.LastMessage.ProjectID,
				Content:    d.LastMessage.Content,
				Timestamp:  d.LastMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
			UnreadCount: d.UnreadCount,
		})
	}

	return &pb.GetDialogsResponse{Dialogs: pbDialogs}, nil
}
