package validate

import (
	"strings"
	"unicode/utf8"

	"gowiki/internal/pkg/apperr"
)

func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return apperr.New(apperr.CodeValidation, 400, field+" 不能为空")
	}
	return nil
}

func Email(value string) error {
	v := strings.TrimSpace(value)
	if v == "" || !strings.Contains(v, "@") || strings.HasPrefix(v, "@") || strings.HasSuffix(v, "@") {
		return apperr.New(apperr.CodeValidation, 400, "邮箱格式不正确")
	}
	return nil
}

func Length(field, value string, min, max int) error {
	n := utf8.RuneCountInString(strings.TrimSpace(value))
	if n < min || n > max {
		return apperr.New(apperr.CodeValidation, 400, field+" 长度不合法")
	}
	return nil
}

func Password(value string) error {
	if utf8.RuneCountInString(value) < 6 {
		return apperr.New(apperr.CodeValidation, 400, "密码至少 6 位")
	}
	return nil
}
