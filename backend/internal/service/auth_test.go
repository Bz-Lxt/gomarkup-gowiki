package service

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"gowiki/internal/config"
	"gowiki/internal/model"
	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/pkg/validate"
)

func TestPasswordAndEmailRules(t *testing.T) {
	if err := validate.Email("bad"); err == nil {
		t.Fatal("email")
	}
	if err := validate.Password("123"); err == nil {
		t.Fatal("password")
	}
	if err := validate.Length("昵称", "", 1, 40); err == nil {
		t.Fatal("name")
	}
}

func TestPickColorStable(t *testing.T) {
	a := pickColor("ada@wiki.dev")
	b := pickColor("ada@wiki.dev")
	if a != b || a == "" {
		t.Fatalf("%s %s", a, b)
	}
}

func TestJWTRoundTrip(t *testing.T) {
	s := &AuthService{cfg: config.Config{JWTSecret: "test-secret", JWTExpireHours: 1}}
	u := &model.User{ID: uuid.New(), Email: "ada@wiki.dev", DisplayName: "Ada", CreatedAt: timeutil.Now()}
	tok, err := s.sign(u, "access", timeutil.Now().Add(time.Hour))
	if err != nil || tok == "" {
		t.Fatal(err)
	}
	c, err := s.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.Email != u.Email || c.TokenType != "access" {
		t.Fatalf("%+v", c)
	}
}
