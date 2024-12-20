package users

import (
	user_dto "chatapp/internal/dto/user"
	"time"
)

type User struct {
	ID        int64     `db:"id" json:"id"`
	Username  string    `db:"user_name" json:"user_name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func FromCreateDto(dto user_dto.CreateUserDTO) *User {
	return &User{
		Username: dto.Name,
		Email:    dto.Email,
		Password: dto.Password,
	}
}
