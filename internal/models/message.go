package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID         string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SenderID   string         `gorm:"type:varchar(255);not null;index"`
	ReceiverID string         `gorm:"type:varchar(255);not null;index"`
	Content    string         `gorm:"type:text;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
