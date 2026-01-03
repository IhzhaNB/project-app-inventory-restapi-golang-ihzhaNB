package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Token     uuid.UUID  `db:"token" json:"token"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

func (s *Session) IsValid() bool {
	now := time.Now()
	return s.RevokedAt == nil && s.ExpiresAt.After(now)
}
