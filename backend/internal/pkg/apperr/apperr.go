package apperr

import "fmt"

type Code string

const (
	CodeBadRequest   Code = "BAD_REQUEST"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeCycle        Code = "TREE_CYCLE"
	CodeLocked       Code = "PARAGRAPH_LOCKED"
	CodeValidation   Code = "VALIDATION"
	CodeInternal     Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	HTTP    int
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

func New(code Code, httpStatus int, msg string) *Error {
	return &Error{Code: code, HTTP: httpStatus, Message: msg}
}

func Wrap(code Code, httpStatus int, msg string, cause error) *Error {
	return &Error{Code: code, HTTP: httpStatus, Message: msg, Cause: cause}
}

var (
	ErrBadRequest   = New(CodeBadRequest, 400, "请求参数不合法")
	ErrUnauthorized = New(CodeUnauthorized, 401, "未登录或令牌已失效")
	ErrForbidden    = New(CodeForbidden, 403, "没有权限执行该操作")
	ErrNotFound     = New(CodeNotFound, 404, "资源不存在")
	ErrConflict     = New(CodeConflict, 409, "资源冲突")
	ErrCycle        = New(CodeCycle, 409, "不能将节点拖入自身子树")
	ErrLocked       = New(CodeLocked, 423, "段落已被他人锁定")
	ErrInternal     = New(CodeInternal, 500, "服务器内部错误")
)
