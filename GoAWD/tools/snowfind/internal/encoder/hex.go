package encoder

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// HexEncoder Hex编码器
type HexEncoder struct {
	BaseEncoder
}

// NewHexEncoder 创建Hex编码器
func NewHexEncoder() *HexEncoder {
	return &HexEncoder{
		BaseEncoder: NewBaseEncoder("hex", "hex编码编码器，尝试识别类似666c6167格式的编码并尝试解码"),
	}
}

// Encode 编码文本为Hex格式
func (e *HexEncoder) Encode(text string) string {
	return hex.EncodeToString([]byte(text))
}

// Decode 解码Hex格式文本
func (e *HexEncoder) Decode(text string) (string, error) {
	decoded, err := hex.DecodeString(text)
	if err != nil {
		return "", fmt.Errorf("hex解码失败: %w", err)
	}
	return string(decoded), nil
}

// GenerateRegex 生成Hex匹配正则表达式
func (e *HexEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	// 取前8个字符作为前缀
	prefix := encoded
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}

	pattern := regexp.QuoteMeta(prefix) + "[0-9a-fA-F]*"
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("生成hex正则表达式失败: %w", err)
	}
	return regex, nil
}

// HexXEncoder Hex编码器（带\x前缀）
type HexXEncoder struct {
	BaseEncoder
}

// NewHexXEncoder 创建HexX编码器
func NewHexXEncoder() *HexXEncoder {
	return &HexXEncoder{
		BaseEncoder: NewBaseEncoder("hex_x", "hex编码编码器，尝试识别类似\\x66\\x6c\\x61\\x67格式的编码并尝试解码"),
	}
}

// Encode 编码文本为HexX格式
func (e *HexXEncoder) Encode(text string) string {
	var result strings.Builder
	for _, b := range []byte(text) {
		result.WriteString(fmt.Sprintf("\\x%02x", b))
	}
	return result.String()
}

// Decode 解码HexX格式文本
func (e *HexXEncoder) Decode(text string) (string, error) {
	// 移除\x前缀
	hexStr := strings.ReplaceAll(text, "\\x", "")
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("hex_x解码失败: %w", err)
	}
	return string(decoded), nil
}

// GenerateRegex 生成HexX匹配正则表达式
func (e *HexXEncoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded) + "(\\\\x[0-9a-fA-F]{2})+"
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("生成hex_x正则表达式失败: %w", err)
	}
	return regex, nil
}
