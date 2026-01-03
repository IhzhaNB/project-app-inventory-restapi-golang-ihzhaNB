package model

type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleStaff      UserRole = "staff"
)

type User struct {
	BaseModel
	Username     string   `db:"username" json:"username"`
	Email        string   `db:"email" json:"email"`
	PasswordHash string   `db:"password_hash" json:"-"`
	FullName     string   `db:"full_name" json:"full_name"`
	Role         UserRole `db:"role" json:"role"`
	IsActive     bool     `db:"is_active" json:"is_active"`
}

// Helper Method
// ============================================
// ROLE CHECKERS
// ============================================

func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsStaff() bool {
	return u.Role == RoleStaff
}

// ============================================
// PERMISSION CHECKERS
// ============================================

// CanManageUsers - untuk user management (admin & super_admin)
func (u *User) CanManageUsers() bool {
	return u.IsSuperAdmin() || u.IsAdmin()
}

// CanManageMasterData - untuk warehouse, category, shelf, product (admin & super_admin)
// Staff tidak boleh create/update/delete master data
func (u *User) CanManageMasterData() bool {
	return u.IsSuperAdmin() || u.IsAdmin()
}

// CanDeleteMasterData - alias untuk consistency (sama dengan CanManageMasterData)
func (u *User) CanDeleteMasterData() bool {
	return u.CanManageMasterData()
}

// CanAccessRevenueReport - hanya admin & super_admin yang bisa lihat report pendapatan
func (u *User) CanAccessRevenueReport() bool {
	return u.IsSuperAdmin() || u.IsAdmin()
}

// ============================================
// SPECIFIC PERMISSION RULES
// ============================================

// CanCreateUserWithRole - cek permission untuk create user dengan role tertentu
// Super admin bisa create semua role, admin hanya bisa create admin & staff
func (u *User) CanCreateUserWithRole(role UserRole) bool {
	if u.IsSuperAdmin() {
		return true // Super admin bisa buat semua role
	}
	if u.IsAdmin() && role != RoleSuperAdmin {
		return true // Admin bisa buat admin & staff, tapi bukan super_admin
	}
	return false // Staff tidak boleh create user
}

// CanUpdateStock - staff boleh update stock (restock/adjustment)
func (u *User) CanUpdateStock() bool {
	// Semua role boleh update stock
	return u.IsSuperAdmin() || u.IsAdmin() || u.IsStaff()
}

// CanCreateSale - staff, admin, super_admin boleh create sale
func (u *User) CanCreateSale() bool {
	return u.IsSuperAdmin() || u.IsAdmin() || u.IsStaff()
}
