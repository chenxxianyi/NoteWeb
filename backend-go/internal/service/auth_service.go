package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(userID uint) error
}

type AuthService struct {
	userRepo  UserRepository
	secretKey string
	expireMin int
}

func NewAuthService(userRepo UserRepository, secretKey string, expireMin int) *AuthService {
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

// ChangePassword changes the user's password after verifying the old password
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// Validate new password
	if len(newPassword) < 6 {
		return errors.New("新密码长度至少6位")
	}

	// Generate new password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Update(user)
}

// UpdateProfile updates user's username and/or email
func (s *AuthService) UpdateProfile(userID uint, username, email string) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// Check if username is taken by another user
	if username != "" && username != user.Username {
		if existing, _ := s.userRepo.GetByUsername(username); existing != nil {
			return nil, errors.New("该用户名已被使用")
		}
		user.Username = username
	}

	// Check if email is taken by another user
	if email != "" && email != user.Email {
		if existing, _ := s.userRepo.GetByEmail(email); existing != nil {
			return nil, errors.New("该邮箱已被注册")
		}
		user.Email = email
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

// UpdateAvatar updates the user's avatar URL
func (s *AuthService) UpdateAvatar(userID uint, avatarURL string) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	user.AvatarURL = avatarURL
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

// DeleteUser deletes a user and all associated data
func (s *AuthService) DeleteUser(userID uint) error {
	// The database should handle cascade deletion
	// For safety, we verify the user exists first
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	return s.userRepo.Delete(user.ID)
}

// VerifyPassword verifies the user's password
func (s *AuthService) VerifyPassword(userID uint, password string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.New("密码错误")
	}

	return nil
}
