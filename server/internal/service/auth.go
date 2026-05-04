package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/model"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthService interface {
	Register(context.Context, *dtov1.UserRegisterReq) (*dtov1.UserRegisterConflictResp, error)
	Login(context.Context, *dtov1.UserLoginReq) (*dtov1.TokenPair, error)
	Refresh(context.Context, string) (*dtov1.TokenPair, error)
	Logout(context.Context, int64) error
	Me(context.Context, int64) (*dtov1.MeResp, error)
}

type authService struct {
	repo repository.UserRepository
	rdb  *redis.Client
	cfg  *config.JwtConfig
}

func NewAuthService(repo repository.UserRepository, rdb *redis.Client, cfg *config.JwtConfig) AuthService {
	return &authService{repo: repo, rdb: rdb, cfg: cfg}
}

func (s *authService) Register(c context.Context, req *dtov1.UserRegisterReq) (*dtov1.UserRegisterConflictResp, error) {
	// 检查 username 或 email 是否被占用
	conflict, err := s.repo.CheckConflict(c, req.Username, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if conflict.IsUsernameConflict || conflict.IsEmailConflict {
		return conflict, consts.ErrRegisterParamConflict
	}

	// 当前邮箱验证码验证服务未实现，先以 000000 未临时替代
	// TODO 邮箱验证码验证服务
	if req.Code != "000000" {
		return nil, consts.ErrInvalidEmailCode
	}

	// BCrypt 加密密码
	hashedByte, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Password: string(hashedByte),
		Email:    req.Email,
		Role:     consts.RoleUser,
	}
	_, err = s.repo.CreateUser(c, user)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *authService) Login(c context.Context, req *dtov1.UserLoginReq) (*dtov1.TokenPair, error) {
	// 检查密码及 identifier 正确性
	user, err := s.repo.FindUserByIdentifier(c, req.Identifier)
	if err != nil {
		// 有可能返回 pgx.ErrNoRows
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		// 有可能返回 bcrypt.ErrMismatchedHashAndPassword
		// TODO 后续可以加上日志审计
		return nil, err
	}

	// 生成 AT 和 RT
	accessToken, err := util.GenerateAccessToken(s.cfg, user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := util.GenerateRefreshToken(s.cfg, user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 将 RT 存入 redis
	key := fmt.Sprintf("%s:%d", consts.RedisRTKey, user.ID)
	if err = s.rdb.Set(c, key, refreshToken, 7*24*time.Hour).Err(); err != nil {
		return nil, err
	}

	// 返回 TokenPair
	resp := &dtov1.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return resp, nil
}

func (s *authService) Refresh(c context.Context, token string) (*dtov1.TokenPair, error) {
	// 校验 RT
	claims, err := util.ParseRefreshToken(s.cfg, token)
	if err != nil {
		return nil, err
	}

	// 查询 redis 的 RT 是否一致
	key := fmt.Sprintf("%s:%d", consts.RedisRTKey, claims.UserID)
	val, err := s.rdb.Get(c, key).Result()
	if err != nil {
		return nil, err
	}
	if val != token {
		return nil, consts.ErrInvalidRefreshToken
	}

	// 签发新的 AT 与 RT
	accessToken, err := util.GenerateAccessToken(s.cfg, claims.UserID, claims.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := util.GenerateRefreshToken(s.cfg, claims.UserID, claims.Username)
	if err != nil {
		return nil, err
	}

	// 将 RT 存入 Redis
	if err = s.rdb.Set(c, key, refreshToken, 7*24*time.Hour).Err(); err != nil {
		return nil, err
	}

	// 返回 TokenPair
	resp := &dtov1.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return resp, nil
}

func (s *authService) Logout(c context.Context, userID int64) error {
	key := fmt.Sprintf("%s:%d", consts.RedisRTKey, userID)
	_, err := s.rdb.Del(c, key).Result()
	return err
}

func (s *authService) Me(c context.Context, userID int64) (*dtov1.MeResp, error) {
	user, err := s.repo.FindUserByID(c, userID)
	if err != nil {
		return nil, err
	}

	resp := &dtov1.MeResp{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	return resp, nil
}
