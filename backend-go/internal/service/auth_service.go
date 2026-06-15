package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	secretKey string
	expireMin int
}

func NewAuthService(userRepo *repository.UserRepo, secretKey string, expireMin int) *AuthService {
	return &AuthService{userRepo: userRepo, secretKey: secretKey, expireMin: expireMin}
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	StorageUsed  int64  `json:"storage_used"`
	StorageLimit int64  `json:"storage_limit"`
}

func toUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID: u.ID, Username: u.Username, Email: u.Email,
		Avatar: u.AvatarURL, StorageUsed: u.StorageUsed, StorageLimit: u.StorageLimit,
	}
}

func (s *AuthService) Register(username, email, password, confirmPassword string) (*AuthResponse, error) {
	if password != confirmPassword {
		return nil, errors.New("密码不一致")
	}
	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, errors.New("该邮箱已被注册")
	}
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return nil, errors.New("该用户名已被使用")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		StorageLimit: 1073741824,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: toUserResponse(user)}, nil
}

func (s *AuthService) Login(email, password string) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: toUserResponse(user)}, nil
}

func (s *AuthService) GetUser(id uint) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func (s *AuthService) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Duration(s.expireMin) * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}
