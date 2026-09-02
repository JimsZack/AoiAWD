package encoder

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// Base64Encoder Base64编码器
type Base64Encoder struct {
	BaseEncoder
}

// NewBase64Encoder 创建Base64编码器
func NewBase64Encoder() *Base64Encoder {
	return &Base64Encoder{
		BaseEncoder: NewBaseEncoder("base64", "Base64编码编码器，尝试识别类似ZmxhZ3sxMjM0fQ==的编码并尝试解码"),
	}
}

// Encode 编码文本为Base64格式
func (e *Base64Encoder) Encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// Decode 解码Base64格式文本
func (e *Base64Encoder) Decode(text string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %w", err)
	}
	return string(decoded), nil
}

// GenerateRegex 生成Base64匹配正则表达式
func (e *Base64Encoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	// 取前4个字符作为前缀
	prefix := encoded
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}

	pattern := regexp.QuoteMeta(prefix) + "(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?"
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("生成base64正则表达式失败: %w", err)
	}
	return regex, nil
}

// URLEncoder URL编码器
type URLEncoder struct {
	BaseEncoder
}

// NewURLEncoder 创建URL编码器
func NewURLEncoder() *URLEncoder {
	return &URLEncoder{
		BaseEncoder: NewBaseEncoder("url", "url编码编码器，尝试识别类似flag%7B123%7D格式的编码并尝试解码"),
	}
}

// Encode 编码文本为URL格式
func (e *URLEncoder) Encode(text string) string {
	var result strings.Builder
	for _, b := range []byte(text) {
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') {
			result.WriteByte(b)
		} else {
			result.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}
	return result.String()
}

// Decode 解码URL格式文本
func (e *URLEncoder) Decode(text string) (string, error) {
	var result strings.Builder
	i := 0
	for i < len(text) {
		if text[i] == '%' && i+2 < len(text) {
			hex := text[i+1 : i+3]
			var b byte
			if _, err := fmt.Sscanf(hex, "%02x", &b); err != nil {
				return "", fmt.Errorf("URL解码失败: %w", err)
			}
			result.WriteByte(b)
			i += 3
		} else {
			result.WriteByte(text[i])
			i++
		}
	}
	return result.String(), nil
}

// GenerateRegex 生成URL匹配正则表达式
func (e *URLEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + "(?:%[0-9A-Fa-f]{2})+"
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("生成URL正则表达式失败: %w", err)
	}
	return regex, nil
}
