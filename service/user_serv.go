package service

import (
	"context"
	"fmt"
	"inventory-system/dto/user"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService interface {
	Create(ctx context.Context, req user.CreateUserRequest) (*user.UserResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*user.UserResponse, error)
	FindAll(ctx context.Context, page int, limit int) ([]user.UserResponse, utils.Pagination, error)
	Update(ctx context.Context, id uuid.UUID, req user.UpdateUserRequest) (*user.UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewUserService(repo *repository.Repository, log *zap.Logger) UserService {
	return &userService{
		repo: repo,
		log:  log,
	}
}

// CREATE USER
// Business logic: validate, hash password, save to db
func (us *userService) Create(ctx context.Context, req user.CreateUserRequest) (*user.UserResponse, error) {
	// 1. Validate input format (pure validation)
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Check email uniqueness (business rule)
	if existing, _ := us.repo.User.FindByEmail(ctx, req.Email); existing != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// 3. Hash password (business logic)
	passwordHash := utils.HashPassword(req.Password)

	// 4. Prepare user object
	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		Role:         model.UserRole(req.Role), // Role sudah divalidasi di handler
		IsActive:     true,
	}

	// 5. Save to database
	if err := us.repo.User.Create(ctx, newUser); err != nil {
		us.log.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user")
	}

	// 6. Return response DTO
	response := us.convertToResponse(newUser)

	us.log.Info("User created", zap.String("user_id", newUser.ID.String()))
	return response, nil
}

// FIND USER BY ID
func (us *userService) FindByID(ctx context.Context, id uuid.UUID) (*user.UserResponse, error) {
	foundUser, err := us.repo.User.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return us.convertToResponse(foundUser), nil
}

// FIND ALL USERS
func (us *userService) FindAll(ctx context.Context, page int, limit int) ([]user.UserResponse, utils.Pagination, error) {
	// Setup pagination
	pagination := utils.NewPagination(page, limit)

	// Get data with pagination
	users, err := us.repo.User.FindAll(ctx, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to get users")
	}

	// Get total count
	total, err := us.repo.User.CountAll(ctx)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count users")
	}

	// Set total in pagination
	pagination.SetTotal(total)

	// Convert to response
	responses := make([]user.UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, *us.convertToResponse(&u))
	}

	us.log.Info("Users fetched with pagination",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int("total", total))

	return responses, pagination, nil
}

// UPDATE USER
// Business logic: validate, update fields
func (us *userService) Update(ctx context.Context, id uuid.UUID, req user.UpdateUserRequest) (*user.UserResponse, error) {
	// Get existing user
	userToUpdate, err := us.repo.User.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	updated := false

	// Update fields if provided and different
	if req.Username != nil && *req.Username != userToUpdate.Username {
		userToUpdate.Username = *req.Username
		updated = true
	}

	if req.Email != nil && *req.Email != userToUpdate.Email {
		userToUpdate.Email = *req.Email
		updated = true
	}

	if req.FullName != nil && *req.FullName != userToUpdate.FullName {
		userToUpdate.FullName = *req.FullName
		updated = true
	}

	// Role sudah divalidasi di handler, kita trust saja
	if req.Role != nil && model.UserRole(*req.Role) != userToUpdate.Role {
		userToUpdate.Role = model.UserRole(*req.Role)
		updated = true
	}

	if req.IsActive != nil && *req.IsActive != userToUpdate.IsActive {
		userToUpdate.IsActive = *req.IsActive
		updated = true
	}

	// Save if changes were made
	if updated {
		if err := us.repo.User.Update(ctx, userToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update user")
		}
	}

	return us.convertToResponse(userToUpdate), nil
}

// DELETE USER
// Business logic: mark as deleted
func (us *userService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := us.repo.User.FindByID(ctx, id); err != nil {
		return fmt.Errorf("user not found")
	}

	if err := us.repo.User.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user")
	}

	us.log.Info("User deleted", zap.String("user_id", id.String()))
	return nil
}

// HELPER Method Response
func (us *userService) convertToResponse(u *model.User) *user.UserResponse {
	return &user.UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
