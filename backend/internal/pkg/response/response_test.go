package response_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
)

func TestFailPreservesWrappedApplicationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	conflict := apperr.New(apperr.CodeConflict, http.StatusConflict, "文档版本已发生变化")
	response.Fail(ctx, fmt.Errorf("save document: %w", conflict))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
	var body response.Body
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != string(apperr.CodeConflict) {
		t.Fatalf("code = %q, want %q", body.Code, apperr.CodeConflict)
	}
	if body.Message != conflict.Message {
		t.Fatalf("message = %q, want %q", body.Message, conflict.Message)
	}
}
