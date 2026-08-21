package validate

import "testing"

func TestHelpers(t *testing.T) {
	if Required("标题", "  ") == nil {
		t.Fatal("required")
	}
	if Email("a@b.com") != nil {
		t.Fatal("good email")
	}
	if Email("@x") == nil || Email("x@") == nil {
		t.Fatal("bad email")
	}
	if Length("t", "你好世界", 1, 2) == nil {
		t.Fatal("too long")
	}
	if Password("123456") != nil {
		t.Fatal("ok pass")
	}
}
