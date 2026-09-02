package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fatih/color"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	logLevelMap = map[string]LogLevel{
		"DEBUG": DEBUG,
		"INFO":  INFO,
		"WARN":  WARN,
		"ERROR": ERROR,
	}
)

type Logger struct {
	level      LogLevel
	infoColor  *color.Color
	warnColor  *color.Color
	errorColor *color.Color
	debugColor *color.Color
	logFile    *os.File
	logWriter  *bufio.Writer
}

var DefaultLogger *Logger

func init() {
	DefaultLogger = &Logger{
		level:      INFO,
		infoColor:  color.New(color.FgGreen),
		warnColor:  color.New(color.FgYellow),
		errorColor: color.New(color.FgRed),
		debugColor: color.New(color.FgCyan),
	}
}

func SetLevel(level string) {
	if l, ok := logLevelMap[level]; ok {
		DefaultLogger.level = l
	}
}

// SetLogFile 设置日志文件
func SetLogFile(filename string) error {
	if DefaultLogger.logFile != nil {
		DefaultLogger.logFile.Close()
	}

	// 确保日志目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}

	DefaultLogger.logFile = file
	DefaultLogger.logWriter = bufio.NewWriter(file)
	return nil
}

// CloseLogFile 关闭日志文件
func CloseLogFile() {
	if DefaultLogger.logWriter != nil {
		DefaultLogger.logWriter.Flush()
	}
	if DefaultLogger.logFile != nil {
		DefaultLogger.logFile.Close()
		DefaultLogger.logFile = nil
		DefaultLogger.logWriter = nil
	}
}

// GenerateLogFileName 生成带时间戳的日志文件名
func GenerateLogFileName() string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("snowfind_log_%s.txt", timestamp)
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(2)

	var levelStr string
	var colorFunc *color.Color

	switch level {
	case DEBUG:
		levelStr = "DEBUG"
		colorFunc = l.debugColor
	case INFO:
		levelStr = "INFO"
		colorFunc = l.infoColor
	case WARN:
		levelStr = "WARN"
		colorFunc = l.warnColor
	case ERROR:
		levelStr = "ERROR"
		colorFunc = l.errorColor
	}

	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s %s:%d - %s",
		levelStr, timestamp, file[max(0, len(file)-20):], line, message)

	// 输出到控制台
	colorFunc.Println(logLine)

	// 同时写入日志文件
	if l.logWriter != nil {
		l.logWriter.WriteString(logLine + "\n")
		l.logWriter.Flush()
	}
}

func Debug(format string, args ...interface{}) {
	DefaultLogger.log(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	DefaultLogger.log(INFO, format, args...)
}

func Warn(format string, args ...interface{}) {
	DefaultLogger.log(WARN, format, args...)
}

func Error(format string, args ...interface{}) {
	DefaultLogger.log(ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	DefaultLogger.log(ERROR, format, args...)
	os.Exit(1)
}

// LogToFile 直接写入日志文件（不输出到控制台）
func LogToFile(format string, args ...interface{}) {
	if DefaultLogger.logWriter != nil {
		message := fmt.Sprintf(format, args...)
		DefaultLogger.logWriter.WriteString(message + "\n")
		DefaultLogger.logWriter.Flush()
	}
}

// 统计信息结构
type Stats struct {
	FilesScanned   int64
	LinesScanned   int64
	MatchesFound   int64
	ErrorsOccurred int64
	ProcessingTime time.Duration
	StartTime      time.Time
	WorkerCount    int
	ProcessCount   int
}

func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
	}
}

func (s *Stats) IncrementFilesScanned() {
	s.FilesScanned++
}

func (s *Stats) IncrementLinesScanned(count int64) {
	s.LinesScanned += count
}

func (s *Stats) IncrementMatches() {
	s.MatchesFound++
}

func (s *Stats) IncrementErrors() {
	s.ErrorsOccurred++
}

func (s *Stats) Finish() {
	s.ProcessingTime = time.Since(s.StartTime)
}

func (s *Stats) PrintSummary() {
	Info("=== 扫描统计信息 ===")
	Info("扫描文件数: %d", s.FilesScanned)
	Info("扫描行数: %d", s.LinesScanned)
	Info("匹配结果: %d", s.MatchesFound)
	Info("错误数量: %d", s.ErrorsOccurred)
	Info("工作线程: %d", s.WorkerCount)
	Info("进程数量: %d", s.ProcessCount)
	Info("处理时间: %.2f秒", s.ProcessingTime.Seconds())
	Info("平均速度: %.2f文件/秒", float64(s.FilesScanned)/s.ProcessingTime.Seconds())
	if s.LinesScanned > 0 {
		Info("行扫描速度: %.2f行/秒", float64(s.LinesScanned)/s.ProcessingTime.Seconds())
	}
	Info("=================")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 进度条功能
type ProgressBar struct {
	total   int64
	current int64
	width   int
	title   string
}

func NewProgressBar(total int64, title string) *ProgressBar {
	return &ProgressBar{
		total: total,
		width: 50,
		title: title,
	}
}

func (p *ProgressBar) Update(current int64) {
	p.current = current
	percent := float64(current) / float64(p.total) * 100
	filled := int(float64(p.width) * percent / 100)

	bar := "["
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += " "
		}
	}
	bar += "]"

	fmt.Printf("\r%s %s %.1f%% (%d/%d)", p.title, bar, percent, current, p.total)
}

func (p *ProgressBar) Finish() {
	p.Update(p.total)
	fmt.Println()
}
