package repository

import (
	"context"
	"fmt"
	"inventory-system/database"
	"inventory-system/model"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionRepo interface {
	Create(ctx context.Context, session *model.Session) error
	FindByToken(ctx context.Context, token uuid.UUID) (*model.Session, error)
	DeleteByToken(ctx context.Context, token uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewSessionRepo(db database.PgxIface, log *zap.Logger) SessionRepo {
	return &sessionRepo{db: db, log: log}
}

// Create - Buat session baru (saat login)
func (sr *sessionRepo) Create(ctx context.Context, session *model.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Generate metadata sebelum insert
	session.ID = uuid.New()
	session.CreatedAt = time.Now()

	// Execute INSERT statement
	_, err := sr.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.Token,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		sr.log.Error("Failed to create session",
			zap.Error(err),
			zap.String("user_id", session.UserID.String()),
		)
		return fmt.Errorf("create session failed: %w", err)
	}

	// Log success untuk audit trail
	sr.log.Info("Session created",
		zap.String("session_id", session.ID.String()),
		zap.String("user_id", session.UserID.String()),
	)
	return nil
}

// FindByToken - Cari session berdasarkan token (untuk validasi)
func (sr *sessionRepo) FindByToken(ctx context.Context, token uuid.UUID) (*model.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, revoked_at, created_at
		FROM sessions 
		WHERE token = $1 AND revoked_at IS NULL
	`

	var session model.Session

	// Query single row berdasarkan token
	err := sr.db.QueryRow(ctx, query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
	)
	if err != nil {
		// Session tidak ditemukan
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Cek apakah session sudah expired
	if time.Now().After(session.ExpiresAt) {
		sr.log.Warn("Session expired",
			zap.String("token", token.String()),
			zap.Time("expires_at", session.ExpiresAt),
		)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// DeleteByToken - Hapus/revoke session berdasarkan token (saat logout)
func (sr *sessionRepo) DeleteByToken(ctx context.Context, token uuid.UUID) error {
	// Delete dengan mengisi token null
	query := `
		UPDATE sessions 
		SET revoked_at = $1 
		WHERE token = $2 AND revoked_at IS NULL
	`

	now := time.Now()
	// Execute delete
	result, err := sr.db.Exec(ctx, query, now, token)
	if err != nil {
		sr.log.Error("Failed to revoke session",
			zap.Error(err),
			zap.String("token", token.String()),
		)
		return fmt.Errorf("revoke session failed: %w", err)
	}

	// Validasi session ditemukan
	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	sr.log.Info("Session revoked",
		zap.String("token", token.String()),
	)
	return nil
}

// DeleteByUserID - Hapus semua session user (force logout semua device)
func (sr *sessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE sessions 
		SET revoked_at = $1 
		WHERE user_id = $2 AND revoked_at IS NULL
	`

	now := time.Now()
	result, err := sr.db.Exec(ctx, query, now, userID)
	if err != nil {
		sr.log.Error("Failed to revoke all user sessions",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("revoke user sessions failed: %w", err)
	}

	sr.log.Info("All user sessions revoked",
		zap.String("user_id", userID.String()),
		zap.Int64("sessions_revoked", result.RowsAffected()),
	)
	return nil
}

// DeleteExpired - Hapus session yang udah expired (cleanup job)
func (sr *sessionRepo) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM sessions 
		WHERE expires_at < $1
	`

	now := time.Now()
	result, err := sr.db.Exec(ctx, query, now)
	if err != nil {
		sr.log.Error("Failed to delete expired sessions",
			zap.Error(err),
		)
		return fmt.Errorf("delete expired sessions failed: %w", err)
	}

	sr.log.Info("Expired sessions cleaned up",
		zap.Int64("sessions_deleted", result.RowsAffected()),
	)
	return nil
}
