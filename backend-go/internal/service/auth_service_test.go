package service

import (
	"errors"
	"testing"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
)

// fakeUserRepo implements UserRepository interface for testing.
type fakeUserRepo struct {
	users        []models.User
	createErr    error
	getByIDErr   error
	updateErr    error
	deleteErr    error
	nextID       uint
}

func (f *fakeUserRepo) GetByID(id uint) (*models.User, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	for i := range f.users {
		if f.users[i].ID == id {
			return &f.users[i], nil
		}
	}
	return nil, errors.New("用户不存在")
}

func (f *fakeUserRepo) GetByEmail(email string) (*models.User, error) {
	for i := range f.users {
		if f.users[i].Email == email {
			return &f.users[i], nil
		}
	}
	return nil, errors.New("用户不存在")
}

func (f *fakeUserRepo) GetByUsername(username string) (*models.User, error) {
	for i := range f.users {
		if f.users[i].Username == username {
			return &f.users[i], nil
		}
	}
	return nil, errors.New("用户不存在")
}

func (f *fakeUserRepo) Create(user *models.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	user.ID = f.nextID
	f.users = append(f.users, *user)
	return nil
}

func (f *fakeUserRepo) Update(user *models.User) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	for i := range f.users {
		if f.users[i].ID == user.ID {
			f.users[i] = *user
			return nil
		}
	}
	return errors.New("用户不存在")
}

func (f *fakeUserRepo) Delete(userID uint) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i := range f.users {
		if f.users[i].ID == userID {
			f.users = append(f.users[:i], f.users[i+1:]...)
			return nil
		}
	}
	return errors.New("用户不存在")
}

func newTestAuthService() *AuthService {
	return NewAuthService(&fakeUserRepo{}, "test-secret", 60)
}

func TestRegister_Success(t *testing.T) {
	svc := newTestAuthService()

	resp, err := svc.Register("testuser", "test@example.com", "Test1234", "Test1234")

	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("Register returned nil response")
	}
	if resp.User.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", resp.User.Username)
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email 'test@example.com', got %q", resp.User.Email)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestRegister_PasswordMismatch(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("testuser", "test@example.com", "Test1234", "Different5678")

	if err == nil {
		t.Fatal("expected error for password mismatch")
	}
	if err.Error() != "密码不一致" {
		t.Fatalf("expected '密码不一致', got %q", err.Error())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestAuthService()

	// First registration succeeds
	_, err := svc.Register("user1", "dup@example.com", "Test1234", "Test1234")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Second with same email should fail
	_, err = svc.Register("user2", "dup@example.com", "Test1234", "Test1234")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if err.Error() != "该邮箱已被注册" {
		t.Fatalf("expected '该邮箱已被注册', got %q", err.Error())
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("dupuser", "a@example.com", "Test1234", "Test1234")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = svc.Register("dupuser", "b@example.com", "Test1234", "Test1234")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	if err.Error() != "该用户名已被使用" {
		t.Fatalf("expected '该用户名已被使用', got %q", err.Error())
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestAuthService()

	// Register first
	_, err := svc.Register("loginuser", "login@example.com", "Pass1234", "Pass1234")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Login with same credentials
	resp, err := svc.Login("login@example.com", "Pass1234")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("Login returned nil response")
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("user", "user@example.com", "CorrectPass1", "CorrectPass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.Login("user@example.com", "WrongPass1")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if err.Error() != "邮箱或密码错误" {
		t.Fatalf("expected '邮箱或密码错误', got %q", err.Error())
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Login("nonexistent@example.com", "SomePass1")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if err.Error() != "邮箱或密码错误" {
		t.Fatalf("expected '邮箱或密码错误', got %q", err.Error())
	}
}

func TestChangePassword_Success(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("chgpw", "chgpw@example.com", "OldPass1", "OldPass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	err = svc.ChangePassword(1, "OldPass1", "NewPass1")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Login with new password should work
	_, err = svc.Login("chgpw@example.com", "NewPass1")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}

	// Login with old password should fail
	_, err = svc.Login("chgpw@example.com", "OldPass1")
	if err == nil {
		t.Fatal("expected error when logging in with old password")
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("user", "user@example.com", "RealPass1", "RealPass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	err = svc.ChangePassword(1, "WrongPass", "NewPass1")
	if err == nil {
		t.Fatal("expected error for wrong old password")
	}
	if err.Error() != "旧密码错误" {
		t.Fatalf("expected '旧密码错误', got %q", err.Error())
	}
}

func TestChangePassword_ShortNewPassword(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("user", "user@example.com", "RealPass1", "RealPass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	err = svc.ChangePassword(1, "RealPass1", "abc")
	if err == nil {
		t.Fatal("expected error for short password")
	}
	if err.Error() != "新密码长度至少6位" {
		t.Fatalf("expected '新密码长度至少6位', got %q", err.Error())
	}
}

func TestGetUser_Success(t *testing.T) {
	svc := newTestAuthService()

	regResp, err := svc.Register("getuser", "get@example.com", "Pass1", "Pass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	userResp, err := svc.GetUser(regResp.User.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if userResp.Username != "getuser" {
		t.Fatalf("expected username 'getuser', got %q", userResp.Username)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.GetUser(999)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("oldname", "old@example.com", "Pass1", "Pass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resp, err := svc.UpdateProfile(1, "newname", "new@example.com")
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if resp.Username != "newname" {
		t.Fatalf("expected username 'newname', got %q", resp.Username)
	}
	if resp.Email != "new@example.com" {
		t.Fatalf("expected email 'new@example.com', got %q", resp.Email)
	}
}

func TestUpdateProfile_DuplicateEmail(t *testing.T) {
	svc := newTestAuthService()

	svc.Register("user1", "user1@example.com", "Pass1", "Pass1")
	svc.Register("user2", "user2@example.com", "Pass1", "Pass1")

	_, err := svc.UpdateProfile(1, "user1", "user2@example.com")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if err.Error() != "该邮箱已被注册" {
		t.Fatalf("expected '该邮箱已被注册', got %q", err.Error())
	}
}

func TestUpdateAvatar_Success(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("avataruser", "avatar@example.com", "Pass1", "Pass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resp, err := svc.UpdateAvatar(1, "/avatars/test.jpg")
	if err != nil {
		t.Fatalf("UpdateAvatar failed: %v", err)
	}
	if resp.Avatar != "/avatars/test.jpg" {
		t.Fatalf("expected avatar '/avatars/test.jpg', got %q", resp.Avatar)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("deluser", "del@example.com", "Pass1", "Pass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	err = svc.DeleteUser(1)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// User should no longer exist
	_, err = svc.GetUser(1)
	if err == nil {
		t.Fatal("expected error after deleting user")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc := newTestAuthService()

	err := svc.DeleteUser(999)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestVerifyPassword_Success(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("verifyuser", "verify@example.com", "MyPass1", "MyPass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	err = svc.VerifyPassword(1, "MyPass1")
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
}

func TestVerifyPassword_Wrong(t *testing.T) {
	svc := newTestAuthService()

	_, err := svc.Register("verifyuser2", "verify2@example.com", "MyPass1", "MyPass1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	err = svc.VerifyPassword(1, "WrongPass")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}
