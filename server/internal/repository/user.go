package repository

import (
	"context"
	"fmt"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type UserRepository interface {
	CreateUser(context.Context, *model.User) (int64, error)
	UpdateUser(context.Context, *dtov1.UserUpdateReq) error
	DeleteUser(context.Context, int64) error
	FindUserByID(context.Context, int64) (*model.User, error)
	FindUserByIdentifier(context.Context, string) (*model.User, error)
	CheckConflict(context.Context, string, string) (*dtov1.UserRegisterConflictResp, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(c context.Context, user *model.User) (int64, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users(username, email, password, role, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, now(), now(), now())
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.Role).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (r *userRepository) UpdateUser(c context.Context, user *dtov1.UserUpdateReq) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	setClauses := []string{}
	args := []any{}
	argIndex := 1

	if user.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *user.Username)
		argIndex++
	}
	if user.Password != nil {
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", argIndex))
		args = append(args, *user.Password)
		argIndex++
	}
	if user.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *user.Email)
		argIndex++
	}
	if user.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *user.Role)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		argIndex,
	)
	args = append(args, user.ID)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *userRepository) DeleteUser(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
		UPDATE users SET deleted_at = NOW() WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *userRepository) FindUserByID(c context.Context, id int64) (*model.User, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
	`

	var resp model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&resp.ID,
		&resp.Username,
		&resp.Email,
		&resp.Password,
		&resp.Role,
		&resp.CreatedAt,
		&resp.UpdatedAt,
		&resp.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (r *userRepository) FindUserByIdentifier(c context.Context, identifier string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password, role, created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 OR email = $1
	`

	var resp model.User
	err := r.db.QueryRow(ctx, query, identifier).Scan(
		&resp.ID,
		&resp.Username,
		&resp.Email,
		&resp.Password,
		&resp.Role,
		&resp.CreatedAt,
		&resp.UpdatedAt,
		&resp.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (r *userRepository) CheckConflict(c context.Context, username, email string) (*dtov1.UserRegisterConflictResp, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	query := `
		SELECT
			EXISTS(SELECT 1 FROM users WHERE username = $1) AS username_exists,
			EXISTS(SELECT 1 FROM users WHERE email = $2) AS email_exists
	`

	var resp dtov1.UserRegisterConflictResp
	err := r.db.QueryRow(ctx, query, username, email).Scan(
		&resp.IsUsernameConflict,
		&resp.IsEmailConflict,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
