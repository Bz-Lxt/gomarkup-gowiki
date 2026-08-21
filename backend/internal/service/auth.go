package service

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"gowiki/internal/config"
	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/pkg/validate"
	"gowiki/internal/repository"
)

type Claims struct {
	UserID      string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"name"`
	TokenType   string `json:"typ"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users *repository.UserRepo
	cfg   config.Config
}

func NewAuthService(users *repository.UserRepo, cfg config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

type AuthResult struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresAt    string     `json:"expiresAt"`
	User         model.User `json:"user"`
}

func (s *AuthService) Register(email, password, name string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := validate.Email(email); err != nil {
		return nil, err
	}
	if err := validate.Password(password); err != nil {
		return nil, err
	}
	if err := validate.Length("昵称", name, 1, 40); err != nil {
		return nil, err
	}
	if _, err := s.users.ByEmail(email); err == nil {
		return nil, apperr.New(apperr.CodeConflict, 409, "邮箱已被注册")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.ErrInternal
	}
	u := &model.User{
		Email: email, PasswordHash: string(hash),
		DisplayName: strings.TrimSpace(name),
		AvatarColor: pickColor(email),
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return s.issue(u)
}

func (s *AuthService) Login(email, password string) (*AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.ByEmail(email)
	if err != nil {
		return nil, apperr.New(apperr.CodeUnauthorized, 401, "邮箱或密码错误")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, apperr.New(apperr.CodeUnauthorized, 401, "邮箱或密码错误")
	}
	return s.issue(u)
}

func (s *AuthService) Refresh(refreshToken string) (*AuthResult, error) {
	claims, err := s.Parse(refreshToken)
	if err != nil || claims.TokenType != "refresh" {
		return nil, apperr.ErrUnauthorized
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, apperr.ErrUnauthorized
	}
	u, err := s.users.ByID(id)
	if err != nil {
		return nil, apperr.ErrUnauthorized
	}
	return s.issue(u)
}

func (s *AuthService) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if parsed == nil || err != nil || !parsed.Valid {
		return nil, apperr.ErrUnauthorized
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, apperr.ErrUnauthorized
	}
	return c, nil
}

func (s *AuthService) issue(u *model.User) (*AuthResult, error) {
	now := timeutil.Now()
	accessExp := now.Add(time.Duration(s.cfg.JWTExpireHours) * time.Hour)
	refreshExp := now.Add(7 * 24 * time.Hour)
	access, err := s.sign(u, "access", accessExp)
	if err != nil {
		return nil, err
	}
	refresh, err := s.sign(u, "refresh", refreshExp)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		AccessToken: access, RefreshToken: refresh,
		ExpiresAt: timeutil.Format(accessExp), User: *u,
	}, nil
}

func (s *AuthService) sign(u *model.User, typ string, exp time.Time) (string, error) {
	now := timeutil.Now()
	c := Claims{
		UserID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName, TokenType: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    "gowiki",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(s.cfg.JWTSecret))
}

func pickColor(seed string) string {
	palette := []string{"#C45C26", "#2F6F4E", "#3D5A80", "#8C4A3A", "#6B4F8A", "#B0892E"}
	sum := 0
	for _, r := range seed {
		sum += int(r)
	}
	return palette[sum%len(palette)]
}
