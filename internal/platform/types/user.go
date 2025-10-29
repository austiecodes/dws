package types

import libdb "github.com/austiecodes/dws/internal/lib/db"

type UserResponse struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

func NewUserResponse(user *libdb.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
	}
}
