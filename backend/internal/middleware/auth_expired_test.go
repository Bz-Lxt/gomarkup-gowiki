package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"gowiki/internal/config"
	"gowiki/internal/middleware"
	"gowiki/internal/service"
)

func TestAuthRejectsExpiredAccessToken(t *testing.T) {
	const secret = "test-secret"
	claims := service.Claims{
		UserID:      uuid.NewString(),
		Email:       "expired@example.com",
		DisplayName: "Expired User",
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			Issuer:    "gowiki",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	auth := service.NewAuthService(nil, config.Config{JWTSecret: secret})
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.Auth(auth))
	called := false
	router.GET("/api/v1/auth/me", func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if called {
		t.Fatal("protected handler ran for an expired access token")
	}
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusUnauthorized, res.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != "UNAUTHORIZED" {
		t.Fatalf("code = %q, want UNAUTHORIZED", body.Code)
	}
}
