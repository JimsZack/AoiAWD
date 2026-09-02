package encoder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// UnicodeEncoder Unicode编码器
type UnicodeEncoder struct {
	BaseEncoder
}

// NewUnicodeEncoder 创建Unicode编码器
func NewUnicodeEncoder() *UnicodeEncoder {
	return &UnicodeEncoder{
		BaseEncoder: NewBaseEncoder("unicode", "unicode编码编码器，尝试识别类似\\u0066\\u006c\\u0061\\u0067格式的编码并尝试解码"),
	}
}

// Encode 编码文本为Unicode格式
func (e *UnicodeEncoder) Encode(text string) string {
	var result strings.Builder
	for _, r := range text {
		result.WriteString(fmt.Sprintf("\\u%04x", r))
	}
	return result.String()
}

// Decode 解码Unicode格式文本
func (e *UnicodeEncoder) Decode(text string) (string, error) {
	// 解析 \uXXXX 格式
	var result strings.Builder
	i := 0
	for i < len(text) {
		if i+5 < len(text) && text[i:i+2] == "\\u" {
			hex := text[i+2 : i+6]
			codePoint, err := strconv.ParseInt(hex, 16, 32)
			if err != nil {
				return "", fmt.Errorf("Unicode解码失败: %w", err)
			}
			result.WriteRune(rune(codePoint))
			i += 6
		} else {
			result.WriteByte(text[i])
			i++
		}
	}
	return result.String(), nil
}

// GenerateRegex 生成Unicode匹配正则表达式
func (e *UnicodeEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + `(?:\\u[0-9A-Fa-f]{4})+`
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("生成Unicode正则表达式失败: %w", err)
	}
	return regex, nil
}

// UnicodeHTMLEntityEncoder Unicode HTML实体编码器
type UnicodeHTMLEntityEncoder struct {
	BaseEncoder
}

// NewUnicodeHTMLEntityEncoder 创建Unicode HTML实体编码器
func NewUnicodeHTMLEntityEncoder() *UnicodeHTMLEntityEncoder {
	return &UnicodeHTMLEntityEncoder{
		BaseEncoder: NewBaseEncoder("unicode_html_entity", "Unicode的HTML实体编码格式编码器，尝试识别类似&#102;&#49;&#97;&#103;格式的编码并尝试解码"),
	}
}

// Encode 编码文本为Unicode HTML实体格式
func (e *UnicodeHTMLEntityEncoder) Encode(text string) string {
	var result strings.Builder
	for _, r := range text {
		result.WriteString(fmt.Sprintf("&#%d;", r))
	}
	return result.String()
}

// Decode 解码Unicode HTML实体格式文本
func (e *UnicodeHTMLEntityEncoder) Decode(text string) (string, error) {
	// 使用正则表达式解析 &#数字; 格式
	re := regexp.MustCompile(`&#(\d+);`)
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		// 提取数字部分
		numStr := re.FindStringSubmatch(match)[1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return match // 如果解析失败，保持原文
		}
		return string(rune(num))
	})
	return result, nil
}

// GenerateRegex 生成Unicode HTML实体匹配正则表达式
func (e *UnicodeHTMLEntityEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + `(?:&#\d+;)+`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成Unicode HTML实体正则表达式失败: %w", err)
	}
	return regex, nil
}
