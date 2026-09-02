package encoder

import (
	"regexp"
)

// PlainTextEncoder 明文编码器
type PlainTextEncoder struct {
	BaseEncoder
}

// NewPlainTextEncoder 创建明文编码器
func NewPlainTextEncoder() *PlainTextEncoder {
	return &PlainTextEncoder{
		BaseEncoder: NewBaseEncoder("plain_text", "明文编码器，用于直接匹配未编码的文本"),
	}
}

// Encode 编码文本（明文直接返回）
func (e *PlainTextEncoder) Encode(text string) string {
	return text
}

// Decode 解码文本（明文直接返回）
func (e *PlainTextEncoder) Decode(text string) (string, error) {
	return text, nil
}

// GenerateRegex 生成明文匹配正则表达式
func (e *PlainTextEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	// 匹配以指定文本开头，后跟任意字符（非贪婪），以}结尾，或者直接匹配文本本身
	pattern := regexp.QuoteMeta(text) + `\{.*?\}|` + regexp.QuoteMeta(text)
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, err
	}
	return regex, nil
}
