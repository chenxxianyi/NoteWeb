package handlers

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
)

// fakeAuthSvc implements AuthHandlerService for testing.
type fakeAuthSvc struct {
	registerResp    *service.AuthResponse
	registerErr     error
	loginResp       *service.AuthResponse
	loginErr        error
	getUserResp     *service.UserResponse
	getUserErr      error
	changePwdErr    error
	updateProfResp  *service.UserResponse
	updateProfErr   error
	updateAvResp    *service.UserResponse
	updateAvErr     error
	verifyPwdErr    error
	deleteUserErr   error
}

func (f *fakeAuthSvc) Register(username, email, password, confirmPassword string) (*service.AuthResponse, error) {
	return f.registerResp, f.registerErr
}
func (f *fakeAuthSvc) Login(email, password string) (*service.AuthResponse, error) {
	return f.loginResp, f.loginErr
}
func (f *fakeAuthSvc) GetUser(id uint) (*service.UserResponse, error) {
	return f.getUserResp, f.getUserErr
}
func (f *fakeAuthSvc) ChangePassword(userID uint, oldPassword, newPassword string) error {
	return f.changePwdErr
}
func (f *fakeAuthSvc) UpdateProfile(userID uint, username, email string) (*service.UserResponse, error) {
	return f.updateProfResp, f.updateProfErr
}
func (f *fakeAuthSvc) UpdateAvatar(userID uint, avatarURL string) (*service.UserResponse, error) {
	return f.updateAvResp, f.updateAvErr
}
func (f *fakeAuthSvc) VerifyPassword(userID uint, password string) error {
	return f.verifyPwdErr
}
func (f *fakeAuthSvc) DeleteUser(userID uint) error {
	return f.deleteUserErr
}

func setupAuthRouter(svc AuthHandlerService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(svc)

	// Unauthenticated routes
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)

	// Authenticated routes (with a middleware that sets userID)
	auth := r.Group("/api/v1/auth")
	auth.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})
	auth.GET("/me", h.Me)
	auth.POST("/change-password", h.ChangePassword)
	auth.PATCH("/profile", h.UpdateProfile)
	auth.POST("/delete-account", h.DeleteAccount)

	return r
}

func checkResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) map[string]interface{} {
	t.Helper()
	if w.Code != expectedCode {
		t.Fatalf("expected status %d, got %d; body: %s", expectedCode, w.Code, w.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error: %v, body: %s", err, w.Body.String())
	}
	return body
}

func TestRegister_Success(t *testing.T) {
	svc := &fakeAuthSvc{
		registerResp: &service.AuthResponse{
			Token: "jwt-token",
			User: service.UserResponse{
				ID: 1, Username: "testuser", Email: "test@example.com",
			},
		},
	}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"username":"testuser","email":"test@example.com","password":"Pass1234","confirm_password":"Pass1234"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 200)
	if resp["token"] != "jwt-token" {
		t.Fatalf("expected token 'jwt-token', got %v", resp["token"])
	}
}

func TestRegister_BadRequest(t *testing.T) {
	router := setupAuthRouter(&fakeAuthSvc{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	checkResponse(t, w, 400)
}

func TestRegister_ServiceError(t *testing.T) {
	svc := &fakeAuthSvc{registerErr: errors.New("该邮箱已被注册")}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"username":"u","email":"dup@test.com","password":"P1","confirm_password":"P1"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 400)
	if resp["detail"] != "该邮箱已被注册" {
		t.Fatalf("expected '该邮箱已被注册', got %v", resp["detail"])
	}
}

func TestLogin_Success(t *testing.T) {
	svc := &fakeAuthSvc{
		loginResp: &service.AuthResponse{
			Token: "login-token",
			User:  service.UserResponse{ID: 1, Username: "user", Email: "user@test.com"},
		},
	}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"email":"user@test.com","password":"Pass1234"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 200)
	if resp["token"] != "login-token" {
		t.Fatalf("expected token 'login-token', got %v", resp["token"])
	}
}

func TestLogin_WrongCredentials(t *testing.T) {
	svc := &fakeAuthSvc{loginErr: errors.New("邮箱或密码错误")}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"email":"bad@test.com","password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 401)
	if resp["detail"] != "邮箱或密码错误" {
		t.Fatalf("expected '邮箱或密码错误', got %v", resp["detail"])
	}
}

func TestMe_Success(t *testing.T) {
	svc := &fakeAuthSvc{
		getUserResp: &service.UserResponse{ID: 1, Username: "meuser", Email: "me@test.com"},
	}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 200)
	if resp["username"] != "meuser" {
		t.Fatalf("expected username 'meuser', got %v", resp["username"])
	}
}

func TestMe_NotFound(t *testing.T) {
	svc := &fakeAuthSvc{getUserErr: errors.New("用户不存在")}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	checkResponse(t, w, 404)
}

func TestChangePassword_Success(t *testing.T) {
	svc := &fakeAuthSvc{}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"old_password":"OldP1","new_password":"NewP1"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 200)
	if resp["detail"] != "密码修改成功" {
		t.Fatalf("expected '密码修改成功', got %v", resp["detail"])
	}
}

func TestChangePassword_Error(t *testing.T) {
	svc := &fakeAuthSvc{changePwdErr: errors.New("旧密码错误")}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"old_password":"Wrong","new_password":"NewP1"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 400)
	if resp["detail"] != "旧密码错误" {
		t.Fatalf("expected '旧密码错误', got %v", resp["detail"])
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	svc := &fakeAuthSvc{
		updateProfResp: &service.UserResponse{ID: 1, Username: "newname", Email: "new@test.com"},
	}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"username":"newname","email":"new@test.com"}`
	req := httptest.NewRequest("PATCH", "/api/v1/auth/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 200)
	if resp["username"] != "newname" {
		t.Fatalf("expected username 'newname', got %v", resp["username"])
	}
}

func TestDeleteAccount_Success(t *testing.T) {
	svc := &fakeAuthSvc{}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"password":"MyPass1"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/delete-account", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 200)
	if resp["detail"] != "账号已删除" {
		t.Fatalf("expected '账号已删除', got %v", resp["detail"])
	}
}

func TestDeleteAccount_WrongPassword(t *testing.T) {
	svc := &fakeAuthSvc{verifyPwdErr: errors.New("密码错误")}
	router := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	body := `{"password":"WrongPass"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/delete-account", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	resp := checkResponse(t, w, 400)
	if resp["detail"] != "密码错误" {
		t.Fatalf("expected '密码错误', got %v", resp["detail"])
	}
}
