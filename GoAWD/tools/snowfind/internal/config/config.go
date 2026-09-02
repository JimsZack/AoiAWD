package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 配置结构
type Config struct {
	Patterns             []string `json:"patterns"`
	MatchKeywords        []string `json:"match"`
	MaxWorkers           int      `json:"max_workers"`            // 最大工作线程数
	MaxProcesses         int      `json:"max_processes"`          // 最大进程数
	LogLevel             string   `json:"log_level"`              // 日志级别
	BufferSize           int      `json:"buffer_size"`            // 缓冲区大小
	ExcludeFiles         []string `json:"exclude_files"`          // 排除文件列表
	ExcludeExts          []string `json:"exclude_exts"`           // 排除文件扩展名
	DefaultPath          string   `json:"default_path"`           // 默认搜索路径
	Interactive          bool     `json:"interactive"`            // 是否启用交互模式
	EnableTimeFilter     bool     `json:"enable_time_filter"`     // 启用时间筛查
	TimeFilterDays       int      `json:"time_filter_days"`       // 时间筛查天数
	SuspiciousExts       []string `json:"suspicious_exts"`        // 可疑文件扩展名
	EnableSuspiciousScan bool     `json:"enable_suspicious_scan"` // 启用可疑文件扫描
}

const ConfigFile = "config.json"

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	// 创建默认配置
	defaultConfig := &Config{
		Patterns:             []string{"flag", "f1ag", "ctfshow", "FLAG", "flag{"},
		MatchKeywords:        []string{"flag{", "flag", "key", "pass"},
		MaxWorkers:           8,
		MaxProcesses:         2,
		LogLevel:             "INFO",
		BufferSize:           8192,
		ExcludeFiles:         []string{"snowfind", "output.txt", "result*.txt"},
		ExcludeExts:          []string{".py", ".pyc", ".exe", ".log"},
		DefaultPath:          ".",
		Interactive:          true,
		EnableTimeFilter:     false,
		TimeFilterDays:       7,
		SuspiciousExts:       []string{".zip", ".rar", ".7z", ".tar", ".gz", ".exe", ".bat", ".sh", ".ps1", ".scr", ".com", ".pif", ".vbs", ".js"},
		EnableSuspiciousScan: false,
	}

	// 如果配置文件不存在，创建默认配置文件
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		fmt.Printf("警告：配置文件 %s 不存在，将创建新的配置文件。\n", ConfigFile)
		if err := SaveConfig(defaultConfig); err != nil {
			return defaultConfig, err
		}
		return defaultConfig, nil
	}

	// 读取现有配置文件
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return defaultConfig, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if len(data) == 0 {
		fmt.Printf("警告：配置文件 %s 为空，将使用默认配置。\n", ConfigFile)
		return defaultConfig, nil
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		fmt.Printf("警告：配置文件 %s 格式不正确，将使用默认配置。\n", ConfigFile)
		return defaultConfig, nil
	}

	// 确保配置不为空
	if len(config.Patterns) == 0 {
		config.Patterns = defaultConfig.Patterns
	}
	if len(config.MatchKeywords) == 0 {
		config.MatchKeywords = defaultConfig.MatchKeywords
	}
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = defaultConfig.MaxWorkers
	}
	if config.MaxProcesses <= 0 {
		config.MaxProcesses = defaultConfig.MaxProcesses
	}
	if config.LogLevel == "" {
		config.LogLevel = defaultConfig.LogLevel
	}
	if config.BufferSize <= 0 {
		config.BufferSize = defaultConfig.BufferSize
	}
	if len(config.ExcludeFiles) == 0 {
		config.ExcludeFiles = defaultConfig.ExcludeFiles
	}
	if len(config.ExcludeExts) == 0 {
		config.ExcludeExts = defaultConfig.ExcludeExts
	}
	if config.DefaultPath == "" {
		config.DefaultPath = defaultConfig.DefaultPath
	}
	if config.TimeFilterDays <= 0 {
		config.TimeFilterDays = defaultConfig.TimeFilterDays
	}
	if len(config.SuspiciousExts) == 0 {
		config.SuspiciousExts = defaultConfig.SuspiciousExts
	}

	return config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(ConfigFile, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

// AddPatterns 添加匹配模式
func (c *Config) AddPatterns(patterns []string) {
	for _, pattern := range patterns {
		if !contains(c.Patterns, pattern) {
			c.Patterns = append(c.Patterns, pattern)
		}
	}
}

// RemovePatterns 删除匹配模式
func (c *Config) RemovePatterns(patterns []string) {
	for _, pattern := range patterns {
		c.Patterns = remove(c.Patterns, pattern)
	}
}

// AddMatchKeywords 添加匹配关键词
func (c *Config) AddMatchKeywords(keywords []string) {
	for _, keyword := range keywords {
		if !contains(c.MatchKeywords, keyword) {
			c.MatchKeywords = append(c.MatchKeywords, keyword)
		}
	}
}

// RemoveMatchKeywords 删除匹配关键词
func (c *Config) RemoveMatchKeywords(keywords []string) {
	for _, keyword := range keywords {
		c.MatchKeywords = remove(c.MatchKeywords, keyword)
	}
}

// 辅助函数：检查切片中是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 辅助函数：从切片中删除某个元素
func remove(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
