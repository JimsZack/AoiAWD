package encoder

import (
	"fmt"
	"regexp"
	"strings"
)

// ROT13Encoder ROT13 编码器
type ROT13Encoder struct{}

// NewROT13Encoder 创建新的 ROT13 编码器
func NewROT13Encoder() Encoder {
	return &ROT13Encoder{}
}

// Name 返回编码器名称
func (e *ROT13Encoder) Name() string {
	return "ROT13"
}

// Description 返回编码器描述
func (e *ROT13Encoder) Description() string {
	return "ROT13 Caesar cipher encoding"
}

// Encode 编码（ROT13 是双向的）
func (e *ROT13Encoder) Encode(input string) string {
	return e.rot13Transform(input)
}

// Decode ROT13 解码（与编码相同）
func (e *ROT13Encoder) Decode(input string) (string, error) {
	return e.rot13Transform(input), nil
}

// IsLikelyEncoded 判断是否可能是 ROT13 编码
func (e *ROT13Encoder) IsLikelyEncoded(input string) bool {
	// ROT13 通常包含字母，且看起来不像正常英文
	hasLetters := regexp.MustCompile(`[a-zA-Z]`).MatchString(input)
	if !hasLetters {
		return false
	}

	// 尝试解码看是否包含常见模式
	decoded := e.rot13Transform(input)
	commonPatterns := []string{"flag", "ctf", "the", "and", "for"}
	for _, pattern := range commonPatterns {
		if strings.Contains(strings.ToLower(decoded), pattern) {
			return true
		}
	}

	return false
}

// Priority 返回优先级
func (e *ROT13Encoder) Priority() int {
	return 60
}

// GenerateRegex 生成正则表达式
func (e *ROT13Encoder) GenerateRegex(input string) (*regexp.Regexp, error) {
	encoded := e.Encode(input)
	pattern := regexp.QuoteMeta(encoded)
	return regexp.Compile(pattern)
}

func (e *ROT13Encoder) rot13Transform(input string) string {
	result := make([]rune, len(input))
	for i, char := range input {
		if char >= 'a' && char <= 'z' {
			result[i] = 'a' + (char-'a'+13)%26
		} else if char >= 'A' && char <= 'Z' {
			result[i] = 'A' + (char-'A'+13)%26
		} else {
			result[i] = char
		}
	}
	return string(result)
}

// AtbashEncoder Atbash 密码编码器
type AtbashEncoder struct{}

// NewAtbashEncoder 创建新的 Atbash 编码器
func NewAtbashEncoder() Encoder {
	return &AtbashEncoder{}
}

// Name 返回编码器名称
func (e *AtbashEncoder) Name() string {
	return "Atbash"
}

// Description 返回编码器描述
func (e *AtbashEncoder) Description() string {
	return "Atbash cipher encoding (A=Z, B=Y, etc.)"
}

// Encode Atbash 编码
func (e *AtbashEncoder) Encode(input string) string {
	return e.atbashTransform(input)
}

// Decode Atbash 解码（与编码相同）
func (e *AtbashEncoder) Decode(input string) (string, error) {
	return e.atbashTransform(input), nil
}

// IsLikelyEncoded 判断是否可能是 Atbash 编码
func (e *AtbashEncoder) IsLikelyEncoded(input string) bool {
	// Atbash 通常包含字母
	hasLetters := regexp.MustCompile(`[a-zA-Z]`).MatchString(input)
	if !hasLetters {
		return false
	}

	// 尝试解码看是否包含常见模式
	decoded := e.atbashTransform(input)
	commonPatterns := []string{"flag", "ctf", "the", "and"}
	for _, pattern := range commonPatterns {
		if strings.Contains(strings.ToLower(decoded), pattern) {
			return true
		}
	}

	return false
}

// Priority 返回优先级
func (e *AtbashEncoder) Priority() int {
	return 65
}

// GenerateRegex 生成正则表达式
func (e *AtbashEncoder) GenerateRegex(input string) (*regexp.Regexp, error) {
	encoded := e.Encode(input)
	pattern := regexp.QuoteMeta(encoded)
	return regexp.Compile(pattern)
}

func (e *AtbashEncoder) atbashTransform(input string) string {
	result := make([]rune, len(input))
	for i, char := range input {
		if char >= 'a' && char <= 'z' {
			result[i] = 'z' - (char - 'a')
		} else if char >= 'A' && char <= 'Z' {
			result[i] = 'Z' - (char - 'A')
		} else {
			result[i] = char
		}
	}
	return string(result)
}

// MorseCodeEncoder 摩尔斯电码编码器
type MorseCodeEncoder struct {
	morseMap        map[rune]string
	reverseMorseMap map[string]rune
}

// NewMorseCodeEncoder 创建新的摩尔斯电码编码器
func NewMorseCodeEncoder() Encoder {
	encoder := &MorseCodeEncoder{
		morseMap: map[rune]string{
			'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".", 'F': "..-.",
			'G': "--.", 'H': "....", 'I': "..", 'J': ".---", 'K': "-.-", 'L': ".-..",
			'M': "--", 'N': "-.", 'O': "---", 'P': ".--.", 'Q': "--.-", 'R': ".-.",
			'S': "...", 'T': "-", 'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-",
			'Y': "-.--", 'Z': "--..",
			'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
			'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
			' ': "/",
		},
	}

	// 创建反向映射
	encoder.reverseMorseMap = make(map[string]rune)
	for char, morse := range encoder.morseMap {
		encoder.reverseMorseMap[morse] = char
	}

	return encoder
}

// Name 返回编码器名称
func (e *MorseCodeEncoder) Name() string {
	return "Morse Code"
}

// Description 返回编码器描述
func (e *MorseCodeEncoder) Description() string {
	return "Morse code encoding using dots and dashes"
}

// Encode 摩尔斯电码编码
func (e *MorseCodeEncoder) Encode(input string) string {
	var result []string
	for _, char := range strings.ToUpper(input) {
		if morse, exists := e.morseMap[char]; exists {
			result = append(result, morse)
		}
	}
	return strings.Join(result, " ")
}

// Decode 摩尔斯电码解码
func (e *MorseCodeEncoder) Decode(input string) (string, error) {
	// 清理输入，移除多余空格
	input = strings.TrimSpace(input)

	// 分割摩尔斯码
	morseCodes := strings.Fields(input)

	var result []rune
	for _, morse := range morseCodes {
		if char, exists := e.reverseMorseMap[morse]; exists {
			result = append(result, char)
		} else {
			// 如果无法解码，返回错误
			return "", fmt.Errorf("invalid morse code: %s", morse)
		}
	}

	return string(result), nil
}

// IsLikelyEncoded 判断是否可能是摩尔斯电码
func (e *MorseCodeEncoder) IsLikelyEncoded(input string) bool {
	// 检查是否主要包含点、短划线和空格
	morsePattern := regexp.MustCompile(`^[.\-\s/]+$`)
	if !morsePattern.MatchString(input) {
		return false
	}

	// 检查是否有合理的分隔
	if strings.Contains(input, "  ") || strings.Contains(input, " ") {
		return true
	}

	// 检查点和短划线的比例
	dots := strings.Count(input, ".")
	dashes := strings.Count(input, "-")

	return (dots > 0 || dashes > 0) && len(input) > 5
}

// Priority 返回优先级
func (e *MorseCodeEncoder) Priority() int {
	return 70
}

// GenerateRegex 生成正则表达式
func (e *MorseCodeEncoder) GenerateRegex(input string) (*regexp.Regexp, error) {
	encoded := e.Encode(input)
	pattern := regexp.QuoteMeta(encoded)
	return regexp.Compile(pattern)
}
