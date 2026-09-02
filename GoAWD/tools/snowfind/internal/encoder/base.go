package encoder

import (
	"fmt"
	"regexp"
)

// Encoder 编码器接口
type Encoder interface {
	// Name 返回编码器名称
	Name() string

	// Description 返回编码器描述
	Description() string

	// Encode 编码文本
	Encode(text string) string

	// Decode 解码文本
	Decode(text string) (string, error)

	// GenerateRegex 生成用于匹配的正则表达式
	GenerateRegex(text string) (*regexp.Regexp, error)
}

// Result 匹配结果
type Result struct {
	FilePath    string `json:"file_path"`
	LineNumber  int    `json:"line_number"`
	LineContent string `json:"line_content"`
	MatchResult string `json:"match_result"`
	DecodedText string `json:"decoded_text,omitempty"`
	MatchFormat string `json:"match_format"`
	EncoderName string `json:"encoder_name"`
}

// String 返回结果的字符串表示
func (r *Result) String() string {
	// 确保显示完整的相对路径
	displayPath := r.FilePath
	if displayPath == "" {
		displayPath = "<未知文件>"
	}

	result := fmt.Sprintf("文件路径: %s\n行号: %d\n匹配行: %s\n匹配结果: %s\n",
		displayPath, r.LineNumber, r.LineContent, r.MatchResult)

	if r.DecodedText != "" {
		result += fmt.Sprintf("尝试解码: %s\n", r.DecodedText)
	}

	result += fmt.Sprintf("匹配格式: %s\n编码器: %s\n", r.MatchFormat, r.EncoderName)
	return result
}

// BaseEncoder 基础编码器，提供通用功能
type BaseEncoder struct {
	name        string
	description string
}

// NewBaseEncoder 创建基础编码器
func NewBaseEncoder(name, description string) BaseEncoder {
	return BaseEncoder{
		name:        name,
		description: description,
	}
}

// Name 返回编码器名称
func (e BaseEncoder) Name() string {
	return e.name
}

// Description 返回编码器描述
func (e BaseEncoder) Description() string {
	return e.description
}
