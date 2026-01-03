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

type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewUserRepo(db database.PgxIface, log *zap.Logger) UserRepo {
	return &userRepo{
		db:  db,
		log: log,
	}
}

func (ur *userRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Generate metadata sebelum insert
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Execute INSERT statement
	_, err := ur.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		ur.log.Error("Failed to create user",
			zap.Error(err),
			zap.String("email", user.Email),
		)
		return fmt.Errorf("Create user Failed: %w", err)
	}

	// Log success untuk audit trail
	ur.log.Info("User Created",
		zap.String("id", user.ID.String()),
		zap.String("email", user.Email),
	)

	return nil
}

func (ur *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active,
		       created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`

	var user model.User

	// Query single row berdasarkan ID
	err := ur.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		// User tidak ditemukan
		return nil, fmt.Errorf("User not found: %w", err)
	}

	return &user, nil
}

func (ur *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active,
		       created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`

	var user model.User

	// Query single row berdasarkan email
	err := ur.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		// User tidak ditemukan
		return nil, fmt.Errorf("User not found: %w", err)
	}

	return &user, nil
}

func (ur *userRepo) FindAll(ctx context.Context) ([]model.User, error) {
	query := `
        SELECT id, username, email, password_hash, full_name, role, is_active,
               created_at, updated_at, deleted_at
        FROM users WHERE deleted_at IS NULL
        ORDER BY created_at DESC
    `

	// Query semua user
	rows, err := ur.db.Query(ctx, query)
	if err != nil {
		ur.log.Error("Failed to query users", zap.Error(err))
		return nil, fmt.Errorf("query users failed: %w", err)
	}
	defer rows.Close()

	// Iterate hasil query
	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			ur.log.Error("Failed to scan user", zap.Error(err))
			return nil, fmt.Errorf("scan user failed: %w", err)
		}

		users = append(users, user)
	}

	// Cek error dari rows
	if err = rows.Err(); err != nil {
		ur.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	ur.log.Info("Fetched all users", zap.Int("total_users", len(users)))
	return users, nil
}

func (ur *userRepo) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password_hash = $3, full_name = $4,
		    role = $5, is_active = $6, updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL
	`

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Execute UPDATE statement
	result, err := ur.db.Exec(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		ur.log.Error("Failed to update user",
			zap.Error(err),
			zap.String("id", user.ID.String()))

		return fmt.Errorf("update user failed: %w", err)
	}

	// Cek jika user benar-benar terupdate
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	ur.log.Info("User updated", zap.String("id", user.ID.String()))
	return nil
}

func (ur *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete dengan mengisi deleted_at
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	now := time.Now()

	// Execute delete
	result, err := ur.db.Exec(ctx, query, now, id)
	if err != nil {
		ur.log.Error("Failed to delete user", zap.Error(err), zap.String("id", id.String()))
		return fmt.Errorf("delete user failed: %w", err)
	}

	// Validasi user ditemukan
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	ur.log.Info("User soft deleted", zap.String("id", id.String()))
	return nil
}
