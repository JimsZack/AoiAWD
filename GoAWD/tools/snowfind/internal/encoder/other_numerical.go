package encoder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// OctalEncoder 八进制编码器
type OctalEncoder struct {
	BaseEncoder
}

// NewOctalEncoder 创建八进制编码器
func NewOctalEncoder() Encoder {
	return &OctalEncoder{
		BaseEncoder: NewBaseEncoder("octal", "八进制编码器，识别如 \\146\\154\\141\\147 格式"),
	}
}

// Encode 编码文本为八进制格式
func (e *OctalEncoder) Encode(text string) string {
	var result strings.Builder
	for _, b := range []byte(text) {
		result.WriteString(fmt.Sprintf("\\%o", b))
	}
	return result.String()
}

// Decode 解码八进制格式文本
func (e *OctalEncoder) Decode(text string) (string, error) {
	parts := strings.Split(text, "\\")
	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		num, err := strconv.ParseInt(part, 8, 32)
		if err != nil {
			return "", fmt.Errorf("八进制解码失败: %w", err)
		}
		result.WriteRune(rune(num))
	}
	return result.String(), nil
}

// GenerateRegex 生成八进制匹配正则表达式
func (e *OctalEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	pattern := `(?:\\(?:[0-7]{1,3})){2,}` // 匹配至少两个八进制序列
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成八进制正则表达式失败: %w", err)
	}
	return regex, nil
}
