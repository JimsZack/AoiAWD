package encoder

import (
	"fmt"
	"regexp"
)

// ReverseEncoder 字符串反转编码器
type ReverseEncoder struct {
	BaseEncoder
}

// NewReverseEncoder 创建字符串反转编码器
func NewReverseEncoder() Encoder {
	return &ReverseEncoder{
		BaseEncoder: NewBaseEncoder("reverse", "字符串反转编码器"),
	}
}

// Encode 编码文本（反转）
func (e *ReverseEncoder) Encode(text string) string {
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Decode 解码文本（反转）
func (e *ReverseEncoder) Decode(text string) (string, error) {
	return e.Encode(text), nil // 反转操作是自反的
}

// GenerateRegex 生成反转匹配正则表达式
func (e *ReverseEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	reversedText := e.Encode(text)
	pattern := regexp.QuoteMeta(reversedText)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成反转正则表达式失败: %w", err)
	}
	return regex, nil
}
