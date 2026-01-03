package user

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=super_admin admin staff"`
}

type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	FullName *string `json:"full_name,omitempty"`
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=super_admin admin staff"`
	IsActive *bool   `json:"is_active,omitempty"`
}
