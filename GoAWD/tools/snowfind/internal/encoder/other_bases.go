package encoder

import (
	"encoding/base32"
	"fmt"
	"regexp"

	"github.com/itchyny/base58-go"
)

// Base32Encoder Base32编码器
type Base32Encoder struct {
	BaseEncoder
}

// NewBase32Encoder 创建Base32编码器
func NewBase32Encoder() Encoder {
	return &Base32Encoder{
		BaseEncoder: NewBaseEncoder("base32", "Base32编码器，使用32个可打印字符（A-Z, 2-7）"),
	}
}

// Encode 编码文本为Base32格式
func (e *Base32Encoder) Encode(text string) string {
	return base32.StdEncoding.EncodeToString([]byte(text))
}

// Decode 解码Base32格式文本
func (e *Base32Encoder) Decode(text string) (string, error) {
	decoded, err := base32.StdEncoding.DecodeString(text)
	if err != nil {
		return "", fmt.Errorf("base32解码失败: %w", err)
	}
	return string(decoded), nil
}

// GenerateRegex 生成Base32匹配正则表达式
func (e *Base32Encoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	// 对输入文本进行 Base32 编码，然后为编码结果生成精确匹配的正则表达式
	encoded := e.Encode(text)
	pattern := regexp.QuoteMeta(encoded)
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("生成base32正则表达式失败: %w", err)
	}
	return regex, nil
}

// Base58Encoder Base58编码器
type Base58Encoder struct {
	BaseEncoder
}

// NewBase58Encoder 创建Base58编码器
func NewBase58Encoder() Encoder {
	return &Base58Encoder{
		BaseEncoder: NewBaseEncoder("base58", "Base58编码器，常用于加密货币地址"),
	}
}

// Encode 编码文本为Base58格式
func (e *Base58Encoder) Encode(text string) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode([]byte(text))
	if err != nil {
		return ""
	}
	return string(encoded)
}

// Decode 解码Base58格式文本
func (e *Base58Encoder) Decode(text string) (string, error) {
	encoding := base58.BitcoinEncoding
	decoded, err := encoding.Decode([]byte(text))
	if err != nil {
		return "", fmt.Errorf("base58解码失败: %w", err)
	}
	return string(decoded), nil
}

// GenerateRegex 生成Base58匹配正则表达式
func (e *Base58Encoder) GenerateRegex(text string) (*regexp.Regexp, error) {
	// 对输入文本进行 Base58 编码，然后为编码结果生成精确匹配的正则表达式
	encoded := e.Encode(text)
	if encoded == "" {
		return nil, fmt.Errorf("base58编码失败")
	}
	pattern := regexp.QuoteMeta(encoded)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("生成base58正则表达式失败: %w", err)
	}
	return regex, nil
}
