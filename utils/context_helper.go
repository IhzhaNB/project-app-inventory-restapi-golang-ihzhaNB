package utils

import (
	"context"
	"inventory-system/model"
)

type contextKey string

const UserContextKey contextKey = "user"

// SetUserToContext digunakan oleh middleware auth untuk menyimpan user
func SetUserToContext(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// GetUserFromContext digunakan oleh handler dan middleware lain untuk mengambil user
func GetUserFromContext(ctx context.Context) *model.User {
	if user, ok := ctx.Value(UserContextKey).(*model.User); ok {
		return user
	}
	return nil
}
