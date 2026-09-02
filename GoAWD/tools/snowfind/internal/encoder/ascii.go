package encoder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ASCIIEncoder ASCII编码器（十进制，用空格分隔）
type ASCIIEncoder struct {
	BaseEncoder
}

// NewASCIIEncoder 创建ASCII编码器
func NewASCIIEncoder() *ASCIIEncoder {
	return &ASCIIEncoder{
		BaseEncoder: NewBaseEncoder("ascii_decimal", "ascii编码十进制编码器，尝试识别类似102 108 97 103格式的编码并尝试解码"),
	}
}

// Encode 编码文本为ASCII十进制格式
func (e *ASCIIEncoder) Encode(text string) string {
	var result []string
	for _, r := range text {
		result = append(result, strconv.Itoa(int(r)))
	}
	return strings.Join(result, " ")
}

// Decode 解码ASCII十进制格式文本
func (e *ASCIIEncoder) Decode(text string) (string, error) {
	parts := strings.Fields(text)
	var result strings.Builder

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return "", fmt.Errorf("ASCII解码失败: %w", err)
		}
		if num < 0 || num > 1114111 { // Unicode范围
			return "", fmt.Errorf("ASCII解码失败: 无效的字符码 %d", num)
		}
		result.WriteRune(rune(num))
	}
	return result.String(), nil
}

// GenerateRegex 生成ASCII匹配正则表达式
func (e *ASCIIEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + ` \d+(?:\s+\d+)*`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成ASCII正则表达式失败: %w", err)
	}
	return regex, nil
}

// ASCIINoSpaceEncoder ASCII编码器（十进制，无空格）
type ASCIINoSpaceEncoder struct {
	BaseEncoder
}

// NewASCIINoSpaceEncoder 创建无空格ASCII编码器
func NewASCIINoSpaceEncoder() *ASCIINoSpaceEncoder {
	return &ASCIINoSpaceEncoder{
		BaseEncoder: NewBaseEncoder("ascii_no_space", "ascii编码编码器，尝试识别类似10210897103123495051125格式的编码并尝试解码"),
	}
}

// Encode 编码文本为ASCII十进制格式（无空格）
func (e *ASCIINoSpaceEncoder) Encode(text string) string {
	var result strings.Builder
	for _, r := range text {
		result.WriteString(strconv.Itoa(int(r)))
	}
	return result.String()
}

// Decode 解码ASCII十进制格式文本（无空格）
func (e *ASCIINoSpaceEncoder) Decode(text string) (string, error) {
	var result strings.Builder

	// 假设每个字符对应3位数字（这是一个简化的假设）
	for i := 0; i < len(text); i += 3 {
		if i+3 > len(text) {
			break
		}
		numStr := text[i : i+3]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return "", fmt.Errorf("ASCII无空格解码失败: %w", err)
		}
		if num < 0 || num > 1114111 {
			return "", fmt.Errorf("ASCII无空格解码失败: 无效的字符码 %d", num)
		}
		result.WriteRune(rune(num))
	}
	return result.String(), nil
}

// GenerateRegex 生成ASCII无空格匹配正则表达式
func (e *ASCIINoSpaceEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + `\d+`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成ASCII无空格正则表达式失败: %w", err)
	}
	return regex, nil
}
