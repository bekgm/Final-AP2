package repository

import (
	"context"

	"github.com/bekgm/Final-AP2/internal/models"
	"gorm.io/gorm"
)

type MessagingRepository interface {
	CreateMessage(ctx context.Context, msg *models.Message) error
	GetMessagesBetweenUsers(ctx context.Context, user1, user2, projectID string, limit, offset int) ([]models.Message, error)
	GetDialogsForUser(ctx context.Context, userID string) ([]DialogResult, error)
}

type messagingRepository struct {
	db *gorm.DB
}

func NewMessagingRepository(db *gorm.DB) MessagingRepository {
	return &messagingRepository{db: db}
}

func (r *messagingRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messagingRepository) GetMessagesBetweenUsers(ctx context.Context, user1, user2, projectID string, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", user1, user2, user2, user1)

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

type DialogResult struct {
	OtherUserID string
	ProjectID   string
	LastMessage models.Message
	UnreadCount int32
}

func (r *messagingRepository) GetDialogsForUser(ctx context.Context, userID string) ([]DialogResult, error) {
	var recentMessages []models.Message
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (
				GREATEST(sender_id, receiver_id),
				LEAST(sender_id, receiver_id),
				project_id
			)
			id, sender_id, receiver_id, project_id, content, created_at, updated_at
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
			ORDER BY 
				GREATEST(sender_id, receiver_id),
				LEAST(sender_id, receiver_id),
				project_id,
				created_at DESC
		`, userID, userID).
		Scan(&recentMessages).Error

	if err != nil {
		return nil, err
	}

	var results []DialogResult
	for _, msg := range recentMessages {
		otherUser := msg.ReceiverID
		if msg.SenderID != userID {
			otherUser = msg.SenderID
		}
		results = append(results, DialogResult{
			OtherUserID: otherUser,
			ProjectID:   msg.ProjectID,
			LastMessage: msg,
			UnreadCount: 0, 
		})
	}

	return results, err
}
