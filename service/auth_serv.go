package service

import (
	"context"
	"fmt"
	"inventory-system/dto/auth"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ============================================
// AUTH SERVICE INTERFACE
// ============================================
type AuthService interface {
	// Login - authenticate user dengan email & password
	Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error)

	// Logout - invalidate session token
	Logout(ctx context.Context, token uuid.UUID) error

	// ValidateToken - cek validitas token dan ambil user data
	ValidateToken(ctx context.Context, token uuid.UUID) (*model.User, error)

	// LogoutAllUserSessions - force logout semua session user (admin feature)
	LogoutAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

// ============================================
// AUTH SERVICE IMPLEMENTATION
// ============================================
type authService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewAuthService(repo *repository.Repository, log *zap.Logger) AuthService {
	return &authService{
		repo: repo,
		log:  log,
	}
}

// ============================================
// LOGIN - AUTHENTICATE USER
// ============================================
// Flow: Validate input → Find user → Check password → Create session → Return token
func (as *authService) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error) {
	// 1. Validate input format
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Find user by email
	user, err := as.repo.User.FindByEmail(ctx, req.Email)
	if err != nil {
		as.log.Warn("Login failed: user not found", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials") // Generic error untuk security
	}

	// 3. Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		as.log.Warn("Login failed: invalid password", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// 4. Check if user is active
	if !user.IsActive {
		as.log.Warn("Login failed: user inactive", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("account is inactive")
	}

	// 5. Generate session token
	token := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour) // Token berlaku 24 jam

	// 6. Create session record
	session := &model.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := as.repo.Session.Create(ctx, session); err != nil {
		as.log.Error("Failed to create session", zap.Error(err), zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("failed to create session")
	}

	// 7. Prepare response
	response := &auth.LoginResponse{
		Token:     token.String(),
		ExpiresAt: expiresAt,
		User: auth.UserInfo{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
			IsActive: user.IsActive,
		},
	}

	as.log.Info("User logged in",
		zap.String("user_id", user.ID.String()),
		zap.String("role", string(user.Role)),
	)

	return response, nil
}

// ============================================
// LOGOUT - INVALIDATE SESSION
// ============================================
// Flow: Parse token → Mark session as revoked
func (as *authService) Logout(ctx context.Context, token uuid.UUID) error {
	// Mark session as revoked (soft delete)
	if err := as.repo.Session.DeleteByToken(ctx, token); err != nil {
		as.log.Error("Failed to logout", zap.Error(err), zap.String("token", token.String()))
		return fmt.Errorf("failed to logout")
	}

	as.log.Info("User logged out", zap.String("token", token.String()))
	return nil
}

// ============================================
// VALIDATE TOKEN - MIDDLEWARE AUTHENTICATION
// ============================================
// Flow: Cek session valid → Cek expired → Cek user aktif
// Digunakan oleh middleware untuk validasi Authorization header
func (as *authService) ValidateToken(ctx context.Context, token uuid.UUID) (*model.User, error) {
	// 1. Find active session by token
	session, err := as.repo.Session.FindByToken(ctx, token)
	if err != nil {
		as.log.Warn("Invalid token", zap.String("token", token.String()), zap.Error(err))
		return nil, fmt.Errorf("invalid or expired token")
	}

	// 2. Get user data from session
	user, err := as.repo.User.FindByID(ctx, session.UserID)
	if err != nil {
		as.log.Error("User not found for valid session",
			zap.String("user_id", session.UserID.String()),
			zap.String("token", token.String()),
		)
		return nil, fmt.Errorf("user not found")
	}

	// 3. Check if user account is active
	if !user.IsActive {
		as.log.Warn("User inactive", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("user account is inactive")
	}

	return user, nil
}

// ============================================
// LOGOUT ALL USER SESSIONS - ADMIN FEATURE
// ============================================
// Force logout semua session user (contoh: saat reset password)
func (as *authService) LogoutAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	if err := as.repo.Session.DeleteByUserID(ctx, userID); err != nil {
		as.log.Error("Failed to logout all sessions",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("failed to logout all sessions")
	}

	as.log.Info("All sessions logged out", zap.String("user_id", userID.String()))
	return nil
}
