package encoder

import (
	"fmt"
	"regexp"
)

// JSFuckEncoder JSFuck编码器
type JSFuckEncoder struct {
	BaseEncoder
}

// NewJSFuckEncoder 创建JSFuck编码器
func NewJSFuckEncoder() Encoder {
	return &JSFuckEncoder{
		BaseEncoder: NewBaseEncoder("jsfuck", "JSFuck编码器，仅使用 []()!+ 六个字符"),
	}
}

// Encode 编码文本为JSFuck格式（仅用于识别，不实现编码）
func (e *JSFuckEncoder) Encode(text string) string {
	return "JSFuck编码无法直接生成"
}

// Decode 解码JSFuck格式文本（仅用于识别，不实现解码）
func (e *JSFuckEncoder) Decode(text string) (string, error) {
	return "JSFuck解码需要JavaScript引擎，暂不支持", nil
}

// GenerateRegex 生成JSFuck匹配正则表达式
func (e *JSFuckEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	// JSFuck 编码过于复杂，暂时返回不匹配任何内容的正则表达式
	// 这样可以避免误匹配
	pattern := `(?!.*)` // 负向前瞻，永远不匹配
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成JSFuck正则表达式失败: %w", err)
	}
	return regex, nil
}
