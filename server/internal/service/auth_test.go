package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/model"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	createUserFn           func(context.Context, *model.User) (int64, error)
	updateUserFn           func(context.Context, *dtov1.UserUpdateReq) error
	deleteUserFn           func(context.Context, int64) error
	findUserByIDFn         func(context.Context, int64) (*model.User, error)
	findUserByIdentifierFn func(context.Context, string) (*model.User, error)
	checkConflictFn        func(context.Context, string, string) (*dtov1.UserRegisterConflictResp, error)

	createdUser *model.User
	updatedUser *dtov1.UserUpdateReq
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *model.User) (int64, error) {
	m.createdUser = user
	if m.createUserFn != nil {
		return m.createUserFn(ctx, user)
	}
	return 2, nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, req *dtov1.UserUpdateReq) error {
	m.updatedUser = req
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, req)
	}
	return nil
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, id int64) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) FindUserByID(ctx context.Context, id int64) (*model.User, error) {
	if m.findUserByIDFn != nil {
		return m.findUserByIDFn(ctx, id)
	}
	return nil, errors.New("unexpected FindUserByID call")
}

func (m *mockUserRepo) FindUserByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	if m.findUserByIdentifierFn != nil {
		return m.findUserByIdentifierFn(ctx, identifier)
	}
	return nil, errors.New("unexpected FindUserByIdentifier call")
}

func (m *mockUserRepo) CheckConflict(ctx context.Context, username, email string) (*dtov1.UserRegisterConflictResp, error) {
	if m.checkConflictFn != nil {
		return m.checkConflictFn(ctx, username, email)
	}
	return &dtov1.UserRegisterConflictResp{}, nil
}

func newTestAuthService(t *testing.T, repo *mockUserRepo) (*authService, *redis.Client, *config.JwtConfig) {
	t.Helper()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
	})

	cfg := &config.JwtConfig{
		AccessSecret:   "test-access-secret",
		RefreshSecret:  "test-refresh-secret",
		AccessExpires:  time.Hour,
		RefreshExpires: 24 * time.Hour,
	}

	return &authService{repo: repo, rdb: rdb, cfg: cfg}, rdb, cfg
}

func TestAuthService_RegisterCreatesUserWithHashedPassword(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{}
	svc, _, _ := newTestAuthService(t, repo)

	_, err := svc.Register(context.Background(), &dtov1.UserRegisterReq{
		Username: "ice",
		Password: "password123",
		Email:    "ice@example.com",
		Code:     "000000",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if repo.createdUser == nil {
		t.Fatal("expected user to be created")
	}
	if repo.createdUser.Role != consts.RoleUser {
		t.Fatalf("expected role %q, got %q", consts.RoleUser, repo.createdUser.Role)
	}
	if repo.createdUser.Password == "password123" {
		t.Fatal("password should be hashed before persistence")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.createdUser.Password), []byte("password123")); err != nil {
		t.Fatalf("created password is not a valid bcrypt hash: %v", err)
	}
	if repo.updatedUser != nil {
		t.Fatal("non-first registered user should not be promoted")
	}
}

func TestAuthService_RegisterFirstUserPromotedToAdmin(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{
		createUserFn: func(context.Context, *model.User) (int64, error) {
			return 1, nil
		},
	}
	svc, _, _ := newTestAuthService(t, repo)

	_, err := svc.Register(context.Background(), &dtov1.UserRegisterReq{
		Username: "ice",
		Password: "password123",
		Email:    "ice@example.com",
		Code:     "000000",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if repo.updatedUser == nil || repo.updatedUser.Role == nil {
		t.Fatal("expected first user role update")
	}
	if *repo.updatedUser.Role != consts.RoleAdmin {
		t.Fatalf("expected first user to be promoted to admin, got %q", *repo.updatedUser.Role)
	}
}

func TestAuthService_RegisterRejectsConflictAndInvalidCode(t *testing.T) {
	t.Parallel()

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		repo := &mockUserRepo{
			checkConflictFn: func(context.Context, string, string) (*dtov1.UserRegisterConflictResp, error) {
				return &dtov1.UserRegisterConflictResp{IsUsernameConflict: true}, nil
			},
		}
		svc, _, _ := newTestAuthService(t, repo)

		conflict, err := svc.Register(context.Background(), &dtov1.UserRegisterReq{
			Username: "ice",
			Password: "password123",
			Email:    "ice@example.com",
			Code:     "000000",
		})
		if !errors.Is(err, consts.ErrRegisterParamConflict) {
			t.Fatalf("expected conflict error, got %v", err)
		}
		if conflict == nil || !conflict.IsUsernameConflict {
			t.Fatalf("expected username conflict data, got %#v", conflict)
		}
		if repo.createdUser != nil {
			t.Fatal("conflicting user should not be created")
		}
	})

	t.Run("invalid email code", func(t *testing.T) {
		t.Parallel()
		repo := &mockUserRepo{}
		svc, _, _ := newTestAuthService(t, repo)

		_, err := svc.Register(context.Background(), &dtov1.UserRegisterReq{
			Username: "ice",
			Password: "password123",
			Email:    "ice@example.com",
			Code:     "123456",
		})
		if !errors.Is(err, consts.ErrInvalidEmailCode) {
			t.Fatalf("expected invalid email code, got %v", err)
		}
		if repo.createdUser != nil {
			t.Fatal("invalid email code should stop before creating user")
		}
	})
}

func TestAuthService_LoginStoresRefreshToken(t *testing.T) {
	t.Parallel()

	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mockUserRepo{
		findUserByIdentifierFn: func(_ context.Context, identifier string) (*model.User, error) {
			if identifier != "ice" {
				t.Fatalf("expected identifier ice, got %q", identifier)
			}
			return &model.User{ID: 42, Username: "ice", Password: string(hashed)}, nil
		},
	}
	svc, rdb, cfg := newTestAuthService(t, repo)

	tokens, err := svc.Login(context.Background(), &dtov1.UserLoginReq{
		Identifier: "ice",
		Password:   "password123",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected token pair, got %#v", tokens)
	}
	if _, err := util.ParseAccessToken(cfg, tokens.AccessToken); err != nil {
		t.Fatalf("access token should parse: %v", err)
	}

	stored, err := rdb.Get(context.Background(), consts.RedisRTKey+":42").Result()
	if err != nil {
		t.Fatalf("refresh token should be stored in redis: %v", err)
	}
	if stored != tokens.RefreshToken {
		t.Fatal("stored refresh token should match response")
	}
}

func TestAuthService_LoginRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mockUserRepo{
		findUserByIdentifierFn: func(context.Context, string) (*model.User, error) {
			return &model.User{ID: 42, Username: "ice", Password: string(hashed)}, nil
		},
	}
	svc, _, _ := newTestAuthService(t, repo)

	_, err = svc.Login(context.Background(), &dtov1.UserLoginReq{
		Identifier: "ice",
		Password:   "wrong-password",
	})
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("expected password mismatch, got %v", err)
	}
}

func TestAuthService_RefreshRotatesRefreshToken(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{}
	svc, rdb, cfg := newTestAuthService(t, repo)
	oldToken, err := util.GenerateRefreshToken(cfg, 42, "ice")
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(context.Background(), consts.RedisRTKey+":42", oldToken, time.Hour).Err(); err != nil {
		t.Fatal(err)
	}

	tokens, err := svc.Refresh(context.Background(), oldToken)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected rotated token pair, got %#v", tokens)
	}
	stored, err := rdb.Get(context.Background(), consts.RedisRTKey+":42").Result()
	if err != nil {
		t.Fatal(err)
	}
	if stored != tokens.RefreshToken {
		t.Fatal("redis should store rotated refresh token")
	}
}

func TestAuthService_RefreshRejectsTokenMismatch(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{}
	svc, rdb, cfg := newTestAuthService(t, repo)
	token, err := util.GenerateRefreshToken(cfg, 42, "ice")
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(context.Background(), consts.RedisRTKey+":42", "other-token", time.Hour).Err(); err != nil {
		t.Fatal(err)
	}

	_, err = svc.Refresh(context.Background(), token)
	if !errors.Is(err, consts.ErrInvalidRefreshToken) {
		t.Fatalf("expected invalid refresh token, got %v", err)
	}
}

func TestAuthService_LogoutDeletesRefreshToken(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{}
	svc, rdb, _ := newTestAuthService(t, repo)
	if err := rdb.Set(context.Background(), consts.RedisRTKey+":42", "rt", time.Hour).Err(); err != nil {
		t.Fatal(err)
	}

	if err := svc.Logout(context.Background(), 42); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
	if exists := rdb.Exists(context.Background(), consts.RedisRTKey+":42").Val(); exists != 0 {
		t.Fatalf("expected refresh token key deleted, exists=%d", exists)
	}
}

func TestAuthService_MeMapsUserResponse(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{
		findUserByIDFn: func(_ context.Context, id int64) (*model.User, error) {
			if id != 42 {
				t.Fatalf("expected id 42, got %d", id)
			}
			return &model.User{ID: 42, Username: "ice", Email: "ice@example.com", Role: consts.RoleAdmin}, nil
		},
	}
	svc, _, _ := newTestAuthService(t, repo)

	resp, err := svc.Me(context.Background(), 42)
	if err != nil {
		t.Fatalf("Me returned error: %v", err)
	}
	if resp.ID != 42 || resp.Username != "ice" || resp.Email != "ice@example.com" || resp.Role != consts.RoleAdmin {
		t.Fatalf("unexpected Me response: %#v", resp)
	}
}
