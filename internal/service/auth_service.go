package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/repository"
	"github.com/kira-app/kira-server/internal/validator"
	"github.com/kira-app/kira-server/pkg/crypto"
	"github.com/kira-app/kira-server/pkg/token"
)

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

type AuthService struct {
	userRepo    *repository.UserRepo
	tokenRepo   *repository.RefreshTokenRepo
	tokenMgr    *token.Manager
}

func NewAuthService(userRepo *repository.UserRepo, tokenRepo *repository.RefreshTokenRepo, tokenMgr *token.Manager) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, tokenMgr: tokenMgr}
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	if !validator.IsEmail(in.Email) {
		return nil, fmt.Errorf("%w: invalid email", domain.ErrBadRequest)
	}
	if !validator.MinLen(in.Name, 2) {
		return nil, fmt.Errorf("%w: name must be at least 2 characters", domain.ErrBadRequest)
	}
	if !validator.MinLen(in.Password, 8) {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", domain.ErrBadRequest)
	}

	existing, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}

	hash, err := crypto.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		ID:              uuid.New(),
		Email:           in.Email,
		Name:            in.Name,
		PasswordHash:    hash,
		DefaultCurrency: domain.CurrencyIDR,
		Timezone:        "Asia/Jakarta",
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, u)
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	u, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if !crypto.CheckPassword(u.PasswordHash, in.Password) {
		return nil, domain.ErrUnauthorized
	}
	return s.issueTokens(ctx, u)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*AuthResult, error) {
	hash := crypto.HashToken(rawToken)
	rt, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}

	if err := s.tokenRepo.Revoke(ctx, hash); err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	hash := crypto.HashToken(rawToken)
	return s.tokenRepo.Revoke(ctx, hash)
}

func (s *AuthService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) UpdateUser(ctx context.Context, u *domain.User) error {
	return s.userRepo.Update(ctx, u)
}

func (s *AuthService) issueTokens(ctx context.Context, u *domain.User) (*AuthResult, error) {
	accessToken, err := s.tokenMgr.NewAccessToken(u.ID, u.Email)
	if err != nil {
		return nil, err
	}

	raw, expiresAt, err := s.tokenMgr.NewRefreshToken()
	if err != nil {
		return nil, err
	}

	hash := crypto.HashToken(raw)
	if err := s.tokenRepo.Create(ctx, u.ID, hash, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: raw,
		User:         u,
	}, nil
}
