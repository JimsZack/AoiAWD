package interactive

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"goawd/tools/snowfind/internal/logger"
	"sort"
	"strconv"
	"strings"
)

// DirectorySelector 目录选择器
type DirectorySelector struct {
	scanner *bufio.Scanner
}

// NewDirectorySelector 创建目录选择器
func NewDirectorySelector() *DirectorySelector {
	return &DirectorySelector{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// SelectPath 交互式选择搜索路径
func (ds *DirectorySelector) SelectPath(defaultPath string) (string, error) {
	logger.Info("启动交互模式，选择搜索路径")

	fmt.Printf("\n=== 搜索路径选择 ===\n")
	fmt.Printf("默认路径: %s\n", defaultPath)

	for {
		fmt.Printf("\n请选择操作:\n")
		fmt.Printf("1. 使用默认路径 (%s)\n", defaultPath)
		fmt.Printf("2. 浏览并选择目录\n")
		fmt.Printf("3. 手动输入路径\n")
		fmt.Printf("4. 显示当前目录内容\n")
		fmt.Printf("请输入选择 (1-4): ")

		if !ds.scanner.Scan() {
			return "", fmt.Errorf("读取用户输入失败")
		}

		choice := strings.TrimSpace(ds.scanner.Text())

		switch choice {
		case "1", "":
			logger.Info("用户选择使用默认路径: %s", defaultPath)
			return defaultPath, nil

		case "2":
			path, err := ds.browseDirectory(defaultPath)
			if err != nil {
				fmt.Printf("浏览目录失败: %v\n", err)
				continue
			}
			return path, nil

		case "3":
			path, err := ds.inputPath()
			if err != nil {
				fmt.Printf("输入路径失败: %v\n", err)
				continue
			}
			return path, nil

		case "4":
			ds.showDirectoryContents(defaultPath)
			continue

		default:
			fmt.Printf("无效选择，请输入 1-4\n")
		}
	}
}

// browseDirectory 浏览目录
func (ds *DirectorySelector) browseDirectory(startPath string) (string, error) {
	currentPath := startPath

	for {
		fmt.Printf("\n当前路径: %s\n", currentPath)

		// 获取目录内容
		dirs, files, err := ds.getDirectoryContents(currentPath)
		if err != nil {
			return "", fmt.Errorf("读取目录失败: %w", err)
		}

		// 显示选项
		fmt.Printf("\n目录和文件:\n")
		options := []string{}

		// 添加上级目录选项
		if currentPath != "/" && currentPath != "." {
			fmt.Printf("0. .. (上级目录)\n")
			options = append(options, "..")
		}

		// 显示目录
		index := 1
		for _, dir := range dirs {
			fmt.Printf("%d. %s/ (目录)\n", index, dir)
			options = append(options, filepath.Join(currentPath, dir))
			index++
		}

		// 显示部分文件（最多5个）
		fileCount := 0
		for _, file := range files {
			if fileCount >= 5 {
				fmt.Printf("   ... (还有 %d 个文件)\n", len(files)-5)
				break
			}
			fmt.Printf("   %s (文件)\n", file)
			fileCount++
		}

		fmt.Printf("\n选择操作:\n")
		fmt.Printf("- 输入数字选择目录\n")
		fmt.Printf("- 输入 'select' 选择当前目录\n")
		fmt.Printf("- 输入 'back' 返回路径选择\n")
		fmt.Printf("请选择: ")

		if !ds.scanner.Scan() {
			return "", fmt.Errorf("读取用户输入失败")
		}

		input := strings.TrimSpace(ds.scanner.Text())

		if input == "select" {
			logger.Info("用户选择目录: %s", currentPath)
			return currentPath, nil
		}

		if input == "back" {
			return ds.SelectPath(startPath)
		}

		// 解析数字选择
		if choice, err := strconv.Atoi(input); err == nil {
			if choice == 0 && len(options) > 0 && options[0] == ".." {
				// 上级目录
				currentPath = filepath.Dir(currentPath)
				if currentPath == "" {
					currentPath = "."
				}
			} else if choice > 0 && choice <= len(dirs) {
				// 选择目录
				if options[0] == ".." {
					currentPath = options[choice]
				} else {
					currentPath = options[choice-1]
				}
			} else {
				fmt.Printf("无效选择，请重新输入\n")
			}
		} else {
			fmt.Printf("无效输入，请输入数字、'select' 或 'back'\n")
		}
	}
}

// inputPath 手动输入路径
func (ds *DirectorySelector) inputPath() (string, error) {
	fmt.Printf("\n请输入搜索路径: ")

	if !ds.scanner.Scan() {
		return "", fmt.Errorf("读取用户输入失败")
	}

	path := strings.TrimSpace(ds.scanner.Text())
	if path == "" {
		return "", fmt.Errorf("路径不能为空")
	}

	// 验证路径是否存在
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("路径不存在或无法访问: %s\n", path)
		fmt.Printf("是否继续使用此路径? (y/N): ")

		if !ds.scanner.Scan() {
			return "", fmt.Errorf("读取确认输入失败")
		}

		confirm := strings.ToLower(strings.TrimSpace(ds.scanner.Text()))
		if confirm != "y" && confirm != "yes" {
			return ds.inputPath()
		}
	}

	logger.Info("用户输入路径: %s", path)
	return path, nil
}

// showDirectoryContents 显示当前目录内容
func (ds *DirectorySelector) showDirectoryContents(path string) {
	dirs, files, err := ds.getDirectoryContents(path)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	fmt.Printf("\n=== 目录内容: %s ===\n", path)

	if len(dirs) > 0 {
		fmt.Printf("目录 (%d个):\n", len(dirs))
		for i, dir := range dirs {
			if i < 10 {
				fmt.Printf("  %s/\n", dir)
			} else if i == 10 {
				fmt.Printf("  ... (还有 %d 个目录)\n", len(dirs)-10)
				break
			}
		}
	}

	if len(files) > 0 {
		fmt.Printf("\n文件 (%d个):\n", len(files))
		for i, file := range files {
			if i < 10 {
				fmt.Printf("  %s\n", file)
			} else if i == 10 {
				fmt.Printf("  ... (还有 %d 个文件)\n", len(files)-10)
				break
			}
		}
	}

	if len(dirs) == 0 && len(files) == 0 {
		fmt.Printf("目录为空\n")
	}
}

// getDirectoryContents 获取目录内容
func (ds *DirectorySelector) getDirectoryContents(path string) ([]string, []string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}

	var dirs, files []string

	for _, entry := range entries {
		name := entry.Name()
		// 跳过隐藏文件
		if strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, name)
		} else {
			files = append(files, name)
		}
	}

	// 排序
	sort.Strings(dirs)
	sort.Strings(files)

	return dirs, files, nil
}

// ConfirmSearch 确认搜索配置
func (ds *DirectorySelector) ConfirmSearch(path string, patterns []string, keywords []string, useKeywordMode bool) bool {
	fmt.Printf("\n=== 搜索配置确认 ===\n")
	fmt.Printf("搜索路径: %s\n", path)

	if useKeywordMode {
		fmt.Printf("搜索模式: 关键词模式\n")
		fmt.Printf("关键词: %v\n", keywords)
	} else {
		fmt.Printf("搜索模式: 编码模式\n")
		fmt.Printf("匹配词: %v\n", patterns)
	}

	fmt.Printf("\n确认开始搜索? (Y/n): ")

	if !ds.scanner.Scan() {
		return false
	}

	confirm := strings.ToLower(strings.TrimSpace(ds.scanner.Text()))
	return confirm == "" || confirm == "y" || confirm == "yes"
}
