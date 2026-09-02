package encoder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BinaryEncoder 二进制编码器（用空格分隔）
type BinaryEncoder struct {
	BaseEncoder
}

// NewBinaryEncoder 创建二进制编码器
func NewBinaryEncoder() *BinaryEncoder {
	return &BinaryEncoder{
		BaseEncoder: NewBaseEncoder("ascii_binary", "ASCII编码二进制编码器，尝试识别类似01100110 01101100 01100001 01100111格式的编码并尝试解码"),
	}
}

// Encode 编码文本为二进制格式（空格分隔）
func (e *BinaryEncoder) Encode(text string) string {
	var result []string
	for _, b := range []byte(text) {
		result = append(result, fmt.Sprintf("%08b", b))
	}
	return strings.Join(result, " ")
}

// Decode 解码二进制格式文本（空格分隔）
func (e *BinaryEncoder) Decode(text string) (string, error) {
	parts := strings.Fields(text)
	var result strings.Builder

	for _, part := range parts {
		num, err := strconv.ParseInt(part, 2, 8)
		if err != nil {
			return "", fmt.Errorf("二进制解码失败: %w", err)
		}
		result.WriteByte(byte(num))
	}
	return result.String(), nil
}

// GenerateRegex 生成二进制匹配正则表达式
func (e *BinaryEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + ` \d+(?:\s+\d+)*`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成二进制正则表达式失败: %w", err)
	}
	return regex, nil
}

// BinaryNoSpaceEncoder 二进制编码器（无空格）
type BinaryNoSpaceEncoder struct {
	BaseEncoder
}

// NewBinaryNoSpaceEncoder 创建无空格二进制编码器
func NewBinaryNoSpaceEncoder() *BinaryNoSpaceEncoder {
	return &BinaryNoSpaceEncoder{
		BaseEncoder: NewBaseEncoder("ascii_binary_no_space", "ASCII编码二进制无空格编码器，尝试识别类似01100110011011000110000101100111格式的编码并尝试解码"),
	}
}

// Encode 编码文本为二进制格式（无空格）
func (e *BinaryNoSpaceEncoder) Encode(text string) string {
	var result strings.Builder
	for _, b := range []byte(text) {
		result.WriteString(fmt.Sprintf("%08b", b))
	}
	return result.String()
}

// Decode 解码二进制格式文本（无空格）
func (e *BinaryNoSpaceEncoder) Decode(text string) (string, error) {
	if len(text)%8 != 0 {
		return "", fmt.Errorf("二进制无空格解码失败: 长度不是8的倍数")
	}

	var result strings.Builder
	for i := 0; i < len(text); i += 8 {
		binaryStr := text[i : i+8]
		num, err := strconv.ParseInt(binaryStr, 2, 8)
		if err != nil {
			return "", fmt.Errorf("二进制无空格解码失败: %w", err)
		}
		result.WriteByte(byte(num))
	}
	return result.String(), nil
}

// GenerateRegex 生成二进制无空格匹配正则表达式
func (e *BinaryNoSpaceEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + `[01]{8}(?:[01]{8})*`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成二进制无空格正则表达式失败: %w", err)
	}
	return regex, nil
}
