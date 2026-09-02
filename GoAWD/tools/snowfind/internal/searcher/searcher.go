package searcher

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"goawd/tools/snowfind/internal/encoder"
	"goawd/tools/snowfind/internal/logger"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Searcher 搜索器
type Searcher struct {
	registry     *encoder.Registry
	workers      int
	stats        *logger.Stats
	excludeFiles []string
	excludeExts  []string
	bufferSize   int
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Patterns             []string // 匹配模式
	MatchKeywords        []string // 关键词匹配
	OutputFile           string   // 输出文件
	UseKeywordMode       bool     // 是否使用关键词模式
	ExcludeFiles         []string // 排除文件列表
	ExcludeExts          []string // 排除文件扩展名
	BufferSize           int      // 文件读取缓冲区大小
	EnableTimeFilter     bool     // 启用时间筛查
	TimeFilterDays       int      // 时间筛查天数
	SuspiciousExts       []string // 可疑文件扩展名
	EnableSuspiciousScan bool     // 启用可疑文件扫描
}

// NewSearcher 创建新的搜索器
func NewSearcher(registry *encoder.Registry, workers int, excludeFiles, excludeExts []string, bufferSize int) *Searcher {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0) // 使用CPU核心数
	}

	stats := logger.NewStats()
	stats.WorkerCount = workers
	stats.ProcessCount = 1 // 当前版本只支持单进程

	logger.Info("初始化搜索器: workers=%d, excludeFiles=%v, excludeExts=%v, bufferSize=%d",
		workers, excludeFiles, excludeExts, bufferSize)

	return &Searcher{
		registry:     registry,
		workers:      workers,
		stats:        stats,
		excludeFiles: excludeFiles,
		excludeExts:  excludeExts,
		bufferSize:   bufferSize,
	}
}

// Search 执行搜索
func (s *Searcher) Search(ctx context.Context, searchPath string, options SearchOptions) ([]*encoder.Result, error) {
	logger.Info("开始搜索: path=%s, keywordMode=%t, timeFilter=%t, suspiciousScan=%t",
		searchPath, options.UseKeywordMode, options.EnableTimeFilter, options.EnableSuspiciousScan)

	// 更新搜索器配置
	if len(options.ExcludeFiles) > 0 {
		s.excludeFiles = options.ExcludeFiles
	}
	if len(options.ExcludeExts) > 0 {
		s.excludeExts = options.ExcludeExts
	}
	if options.BufferSize > 0 {
		s.bufferSize = options.BufferSize
	}

	files, err := s.getFiles(searchPath, options)
	if err != nil {
		logger.Error("获取文件列表失败: %v", err)
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}

	if len(files) == 0 {
		logger.Warn("没有找到可搜索的文件")
		return nil, fmt.Errorf("没有找到可搜索的文件")
	}

	logger.Info("找到 %d 个文件待扫描", len(files))

	// 1. 可疑文件检测（独立功能）
	if options.EnableSuspiciousScan {
		s.performSuspiciousFileScan(files, options)
	}

	// 2. 关键词搜索（独立功能）
	var results []*encoder.Result
	hasKeywordSearch := (options.UseKeywordMode && len(options.MatchKeywords) > 0) ||
		(!options.UseKeywordMode && len(options.Patterns) > 0)

	if hasKeywordSearch {
		// 创建进度条
		progress := logger.NewProgressBar(int64(len(files)), "扫描进度")

		if options.UseKeywordMode {
			// 关键词模式搜索
			results, err = s.searchKeywords(ctx, files, options.MatchKeywords, progress)
		} else {
			// 编码模式搜索
			results, err = s.searchWithEncoders(ctx, files, options.Patterns, progress)
		}

		progress.Finish()

		if err != nil {
			logger.Error("搜索过程中发生错误: %v", err)
			return nil, err
		}

		logger.Info("关键词搜索完成，找到 %d 个匹配结果", len(results))
	} else {
		logger.Info("未启用关键词搜索，跳过内容搜索")
	}

	s.stats.Finish()
	logger.Info("搜索完成，找到 %d 个匹配结果", len(results))

	// 将匹配结果写入日志文件
	if len(results) > 0 {
		logger.LogToFile("\n=== 匹配结果详情 ===")
		for i, result := range results {
			logger.LogToFile("结果 %d:", i+1)
			logger.LogToFile("%s", result.String())
		}
		logger.LogToFile("=== 匹配结果结束 ===\n")
	}

	// 输出到文件
	if options.OutputFile != "" {
		outputPath := s.generateOutputPath(options.OutputFile)
		logger.Info("开始写入结果到文件: %s", outputPath)
		if err := s.writeResultsToFile(results, outputPath); err != nil {
			logger.Error("写入输出文件失败: %v", err)
			return results, fmt.Errorf("写入输出文件失败: %w", err)
		}
		logger.Info("结果已写入文件: %s", outputPath)
	}

	// 打印统计信息
	s.stats.PrintSummary()

	return results, nil
}

// searchKeywords 关键词模式搜索
func (s *Searcher) searchKeywords(ctx context.Context, files []string, keywords []string, progress *logger.ProgressBar) ([]*encoder.Result, error) {
	if len(keywords) == 0 {
		return nil, fmt.Errorf("关键词列表为空")
	}

	logger.Info("开始关键词搜索: keywords=%v", keywords)

	// 构建关键词正则表达式
	pattern := strings.Join(keywords, "|")
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("编译关键词正则表达式失败: %w", err)
	}

	return s.searchFilesWithRegex(ctx, files, regex, "keyword_match", "None", progress)
}

// searchWithEncoders 使用编码器搜索
func (s *Searcher) searchWithEncoders(ctx context.Context, files []string, patterns []string, progress *logger.ProgressBar) ([]*encoder.Result, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("匹配模式列表为空")
	}

	logger.Info("开始编码器搜索: patterns=%v", patterns)

	var results []*encoder.Result
	var mu sync.Mutex
	var processedFiles int64

	// 创建工作队列
	type task struct {
		filePath string
		pattern  string
		encoder  encoder.Encoder
	}

	totalTasks := len(files) * len(patterns) * len(s.registry.GetEncoders())
	logger.Debug("总任务数: %d", totalTasks)

	taskChan := make(chan task, totalTasks)
	resultChan := make(chan []*encoder.Result, s.workers*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			logger.Debug("启动工作协程 %d", workerID)
			for t := range taskChan {
				select {
				case <-ctx.Done():
					logger.Debug("工作协程 %d 接收到取消信号", workerID)
					return
				default:
					fileResults := s.searchFileWithEncoder(t.filePath, t.pattern, t.encoder)
					if len(fileResults) > 0 {
						select {
						case resultChan <- fileResults:
						case <-ctx.Done():
							return
						}
					}
					// 更新进度
					if atomic.AddInt64(&processedFiles, 1)%int64(len(patterns)*len(s.registry.GetEncoders())) == 0 {
						progress.Update(atomic.LoadInt64(&processedFiles) / int64(len(patterns)*len(s.registry.GetEncoders())))
					}
				}
			}
		}(i)
	}

	// 收集结果
	go func() {
		for fileResults := range resultChan {
			mu.Lock()
			results = append(results, fileResults...)
			s.stats.IncrementMatches()
			mu.Unlock()
		}
	}()

	// 生成任务
	go func() {
		defer close(taskChan)
		for _, filePath := range files {
			for _, pattern := range patterns {
				for _, enc := range s.registry.GetEncoders() {
					select {
					case <-ctx.Done():
						logger.Debug("任务生成器接收到取消信号")
						return
					case taskChan <- task{filePath: filePath, pattern: pattern, encoder: enc}:
					}
				}
			}
		}
	}()

	// 等待所有工作完成
	wg.Wait()
	close(resultChan)

	logger.Info("编码器搜索完成，找到 %d 个结果", len(results))
	return results, nil
}

// searchFileWithEncoder 使用指定编码器搜索单个文件
func (s *Searcher) searchFileWithEncoder(filePath, pattern string, enc encoder.Encoder) []*encoder.Result {
	regex, err := enc.GenerateRegex(pattern)
	if err != nil {
		logger.Debug("生成正则表达式失败: %v", err)
		return nil
	}

	results, err := s.searchFilesWithRegex(context.Background(), []string{filePath}, regex, enc.Encode(pattern), enc.Name(), nil)
	if err != nil {
		logger.Debug("搜索文件失败: %v", err)
		return nil
	}

	// 尝试解码匹配结果
	for _, result := range results {
		if decoded, err := enc.Decode(result.MatchResult); err == nil {
			result.DecodedText = decoded
		}
	}

	return results
}

// searchFilesWithRegex 使用正则表达式搜索文件
func (s *Searcher) searchFilesWithRegex(ctx context.Context, files []string, regex *regexp.Regexp, matchFormat, encoderName string, progress *logger.ProgressBar) ([]*encoder.Result, error) {
	var results []*encoder.Result
	var processedFiles int64

	for _, filePath := range files {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		logger.Debug("搜索文件: %s", filePath)
		fileResults, err := s.searchSingleFile(filePath, regex, matchFormat, encoderName)
		if err != nil {
			logger.Debug("跳过文件 %s: %v", filePath, err)
			s.stats.IncrementErrors()
			continue // 跳过无法读取的文件
		}

		if len(fileResults) > 0 {
			results = append(results, fileResults...)
			logger.Debug("文件 %s 找到 %d 个匹配", filePath, len(fileResults))
		}

		s.stats.IncrementFilesScanned()

		// 更新进度
		current := atomic.AddInt64(&processedFiles, 1)
		if progress != nil {
			progress.Update(current)
		}
	}

	return results, nil
}

// searchSingleFile 搜索单个文件
func (s *Searcher) searchSingleFile(filePath string, regex *regexp.Regexp, matchFormat, encoderName string) ([]*encoder.Result, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var results []*encoder.Result
	scanner := bufio.NewScanner(file)

	// 设置缓冲区大小
	if s.bufferSize > 0 {
		buffer := make([]byte, s.bufferSize)
		scanner.Buffer(buffer, s.bufferSize)
	}

	lineNumber := 1
	linesScanned := int64(0)

	for scanner.Scan() {
		line := scanner.Text()
		linesScanned++

		// 过滤可打印字符
		filteredLine := s.filterPrintableChars(line)

		matches := regex.FindAllString(filteredLine, -1)
		for _, match := range matches {
			// 清理和标准化文件路径显示
			displayPath := s.cleanFilePath(filePath)

			result := &encoder.Result{
				FilePath:    displayPath,
				LineNumber:  lineNumber,
				LineContent: strings.TrimSpace(filteredLine),
				MatchResult: match,
				MatchFormat: matchFormat,
				EncoderName: encoderName,
			}
			results = append(results, result)
			logger.Debug("找到匹配: %s 在 %s:%d", match, filePath, lineNumber)
		}
		lineNumber++
	}

	s.stats.IncrementLinesScanned(linesScanned)

	if err := scanner.Err(); err != nil {
		return results, fmt.Errorf("读取文件内容失败: %w", err)
	}

	return results, nil
}

// filterPrintableChars 过滤可打印字符
func (s *Searcher) filterPrintableChars(input string) string {
	var result strings.Builder
	for _, r := range input {
		if r >= 32 && r <= 126 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// cleanFilePath 清理和标准化文件路径显示
func (s *Searcher) cleanFilePath(filePath string) string {
	// 获取当前工作目录
	pwd, err := os.Getwd()
	if err != nil {
		return filePath // 如果无法获取工作目录，返回原路径
	}

	// 尝试获取相对于当前目录的路径
	relPath, err := filepath.Rel(pwd, filePath)
	if err != nil {
		return filePath // 如果无法获取相对路径，返回原路径
	}

	// 如果相对路径更短且不以 ".." 开头，使用相对路径
	if len(relPath) < len(filePath) && !strings.HasPrefix(relPath, "..") {
		return relPath
	}

	return filePath
}

// getFiles 获取所有需要搜索的文件
func (s *Searcher) getFiles(searchPath string, options SearchOptions) ([]string, error) {
	var files []string

	info, err := os.Stat(searchPath)
	if err != nil {
		return nil, fmt.Errorf("访问路径失败: %w", err)
	}

	if info.IsDir() {
		// 遍历目录
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 跳过无法访问的文件
			}

			if !info.IsDir() && s.shouldIncludeFile(path, options) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("遍历目录失败: %w", err)
		}
	} else {
		// 单个文件
		if s.shouldIncludeFile(searchPath, options) {
			files = append(files, searchPath)
		}
	}

	return files, nil
}

// shouldIncludeFile 判断是否应该包含该文件
func (s *Searcher) shouldIncludeFile(filePath string, options SearchOptions) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	fileName := strings.ToLower(filepath.Base(filePath))

	// 检查排除的文件扩展名
	for _, excludeExt := range s.excludeExts {
		if ext == excludeExt {
			logger.Debug("跳过文件 (扩展名): %s", filePath)
			return false
		}
	}

	// 检查排除的文件名模式
	for _, excludeFile := range s.excludeFiles {
		if matched, _ := filepath.Match(excludeFile, fileName); matched {
			logger.Debug("跳过文件 (文件名匹配): %s", filePath)
			return false
		}
		if strings.Contains(fileName, excludeFile) {
			logger.Debug("跳过文件 (文件名包含): %s", filePath)
			return false
		}
	}

	// 时间筛查
	if options.EnableTimeFilter && options.TimeFilterDays > 0 {
		if !s.isFileWithinTimeRange(filePath, options.TimeFilterDays) {
			logger.Debug("跳过文件 (时间筛查): %s", filePath)
			return false
		}
	}

	// 跳过二进制文件（简单检测）
	if s.isBinaryFile(filePath) {
		logger.Debug("跳过文件 (二进制文件): %s", filePath)
		return false
	}

	return true
}

// isBinaryFile 简单检测是否为二进制文件
func (s *Searcher) isBinaryFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// 读取文件头部分进行检测
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return false
	}

	// 检查是否包含过多的null字节
	nullCount := 0
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			nullCount++
		}
	}

	// 如果null字节超过20%，认为是二进制文件
	return float64(nullCount)/float64(n) > 0.2
}

// generateOutputPath 生成带时间戳的输出路径
func (s *Searcher) generateOutputPath(outputFile string) string {
	if outputFile == "" {
		return ""
	}

	// 获取文件扩展名
	ext := filepath.Ext(outputFile)
	name := strings.TrimSuffix(outputFile, ext)

	// 生成时间戳
	timestamp := time.Now().Format("20060102_150405")

	// 组合新的文件名
	return fmt.Sprintf("%s_output_%s%s", name, timestamp, ext)
}

// writeResultsToFile 将结果写入文件
func (s *Searcher) writeResultsToFile(results []*encoder.Result, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, result := range results {
		if _, err := writer.WriteString(result.String() + "\n"); err != nil {
			return fmt.Errorf("写入结果失败: %w", err)
		}
	}

	return nil
}

// performSuspiciousFileScan 执行可疑文件扫描
func (s *Searcher) performSuspiciousFileScan(files []string, options SearchOptions) {
	logger.Info("开始可疑文件检测...")

	suspiciousFiles := s.findSuspiciousFiles(files, options)

	if len(suspiciousFiles) > 0 {
		logger.Warn("发现 %d 个可疑文件:", len(suspiciousFiles))
		for _, file := range suspiciousFiles {
			logger.Warn("  - %s", file)
		}

		// 输出可疑文件报告到文件
		if options.OutputFile != "" {
			outputPath := s.generateSuspiciousReportPath(options.OutputFile)
			if err := s.writeSuspiciousFilesToFile(suspiciousFiles, outputPath); err != nil {
				logger.Error("写入可疑文件报告失败: %v", err)
			} else {
				logger.Info("可疑文件报告已写入: %s", outputPath)
			}
		}
	} else {
		logger.Info("未发现可疑文件")
	}
}

// findSuspiciousFiles 查找可疑文件
func (s *Searcher) findSuspiciousFiles(files []string, options SearchOptions) []string {
	if !options.EnableSuspiciousScan || len(options.SuspiciousExts) == 0 {
		return nil
	}

	var suspiciousFiles []string
	cutoffTime := time.Now().AddDate(0, 0, -options.TimeFilterDays)

	for _, filePath := range files {
		ext := strings.ToLower(filepath.Ext(filePath))

		// 检查是否为可疑扩展名
		isSuspicious := false
		for _, suspExt := range options.SuspiciousExts {
			if ext == strings.ToLower(suspExt) {
				isSuspicious = true
				break
			}
		}

		if !isSuspicious {
			continue
		}

		// 获取文件信息
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		// 检查文件修改时间
		if options.EnableTimeFilter && options.TimeFilterDays > 0 {
			if fileInfo.ModTime().After(cutoffTime) {
				suspiciousFiles = append(suspiciousFiles, fmt.Sprintf("%s (修改时间: %s, 大小: %d bytes)",
					filePath, fileInfo.ModTime().Format("2006-01-02 15:04:05"), fileInfo.Size()))
			}
		} else {
			// 如果不启用时间筛查，仍然报告可疑文件
			suspiciousFiles = append(suspiciousFiles, fmt.Sprintf("%s (修改时间: %s, 大小: %d bytes)",
				filePath, fileInfo.ModTime().Format("2006-01-02 15:04:05"), fileInfo.Size()))
		}
	}

	return suspiciousFiles
}

// isFileWithinTimeRange 检查文件是否在指定时间范围内被修改
func (s *Searcher) isFileWithinTimeRange(filePath string, days int) bool {
	if days <= 0 {
		return true // 如果没有设置天数限制，包含所有文件
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false // 无法获取文件信息，不包含该文件
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	return fileInfo.ModTime().After(cutoffTime)
}

// generateSuspiciousReportPath 生成可疑文件报告路径
func (s *Searcher) generateSuspiciousReportPath(outputFile string) string {
	if outputFile == "" {
		return ""
	}

	// 获取文件扩展名
	ext := filepath.Ext(outputFile)
	name := strings.TrimSuffix(outputFile, ext)

	// 生成时间戳
	timestamp := time.Now().Format("20060102_150405")

	// 组合新的文件名
	return fmt.Sprintf("%s_suspicious_%s%s", name, timestamp, ext)
}

// writeSuspiciousFilesToFile 将可疑文件列表写入文件
func (s *Searcher) writeSuspiciousFilesToFile(suspiciousFiles []string, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("输出文件路径为空")
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建可疑文件报告失败: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 写入报告头部
	header := fmt.Sprintf("=== 可疑文件检测报告 ===\n生成时间: %s\n文件总数: %d\n\n",
		time.Now().Format("2006-01-02 15:04:05"), len(suspiciousFiles))
	if _, err := writer.WriteString(header); err != nil {
		return fmt.Errorf("写入报告头部失败: %w", err)
	}

	// 写入可疑文件列表
	for i, suspiciousFile := range suspiciousFiles {
		line := fmt.Sprintf("%d. %s\n", i+1, suspiciousFile)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("写入可疑文件信息失败: %w", err)
		}
	}

	return nil
}
