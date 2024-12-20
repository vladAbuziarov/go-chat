package user

import (
	"chatapp/internal/constants"
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/users"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, userDto *user_dto.CreateUserDTO) (*users.User, error) {
	user := &users.User{
		Username:  userDto.Name,
		Email:     userDto.Email,
		Password:  userDto.Password,
		CreatedAt: time.Now(),
	}

	query, args, err := r.db.BindNamed(fmt.Sprintf(`INSERT INTO %s (user_name, email, password, created_at) VALUES (:user_name, :email, :password, :created_at) RETURNING id`, constants.UserTable), user)
	if err != nil {
		return nil, err
	}
	if err = r.db.GetContext(ctx, &user.ID, query, args...); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	query := fmt.Sprintf("select * from %s where email = $1 limit 1", constants.UserTable)
	user := &users.User{}
	if err := r.db.GetContext(ctx, user, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
func (r *Repository) GetUserById(ctx context.Context, id int64) (*users.User, error) {
	query := fmt.Sprintf("select * from %s where id = $1 limit 1", constants.UserTable)
	user := &users.User{}
	if err := r.db.GetContext(ctx, user, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
