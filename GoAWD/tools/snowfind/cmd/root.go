package cmd

import (
	"context"
	"fmt"
	"os"
	"goawd/tools/snowfind/internal/config"
	"goawd/tools/snowfind/internal/encoder"
	"goawd/tools/snowfind/internal/interactive"
	"goawd/tools/snowfind/internal/logger"
	"goawd/tools/snowfind/internal/searcher"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	outputFile         string
	patterns           []string
	addPatterns        []string
	deletePatterns     []string
	showEncoders       bool
	useKeywordMode     bool
	addKeywords        []string
	deleteKeywords     []string
	workers            int
	interactiveMode    bool
	nonInteractive     bool
	enableTimeFilter   bool
	timeFilterDays     int
	enableSuspicious   bool
	suspiciousExtsList []string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "snowfind [path]",
	Short: "snowfind - CTF比赛中自动查找FLAG的工具",
	Long: `snowfind是一款自动查找flag并尝试自动解密的工具，
可以尝试通过关键字一键找出flag，在流量分析或者隐写中可以尝试一键梭哈。

功能特色：
• 自动识别单文件与文件夹，当路径是文件时自动分析文件，当路径是文件夹时将自动分析文件夹下所有文件
• 支持交互式目录选择，方便用户浏览和选择目标路径
• 支持自动将关键字编码为多种编码并进行匹配
• 支持匹配关键词后尝试自动解码
• 支持自定义编码解码
• 支持时间筛查，可以检测指定时间内新增或修改的文件
• 支持可疑文件检测，自动识别压缩包、可执行文件等潜在威胁
• 使用Go语言重写，性能更优，支持并发处理`,
	Args: cobra.MaximumNArgs(1),
	Run:  run,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "指定输出文件路径")
	rootCmd.Flags().StringSliceVarP(&patterns, "pattern", "p", nil, "指定匹配词，指定匹配词后将使用当前匹配词不再读取配置文件")
	rootCmd.Flags().StringSliceVar(&addPatterns, "add", nil, "添加新的匹配词到配置文件")
	rootCmd.Flags().StringSliceVar(&deletePatterns, "del", nil, "删除指定的匹配词")
	rootCmd.Flags().BoolVar(&showEncoders, "show-encoders", false, "显示已加载的编码器")
	rootCmd.Flags().BoolVarP(&useKeywordMode, "match", "m", false, "以匹配关键字搜索，该模式不再尝试解码")
	rootCmd.Flags().StringSliceVar(&addKeywords, "add-match", nil, "添加新的匹配关键字到配置文件")
	rootCmd.Flags().StringSliceVar(&deleteKeywords, "del-match", nil, "从配置文件中删除指定的匹配关键字")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 4, "并发工作协程数")
	rootCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "启用交互模式选择搜索路径")
	rootCmd.Flags().BoolVar(&nonInteractive, "no-interactive", false, "禁用交互模式，使用默认路径")
	rootCmd.Flags().BoolVar(&enableTimeFilter, "time-filter", false, "启用时间筛查，只扫描指定天数内修改的文件")
	rootCmd.Flags().IntVar(&timeFilterDays, "days", 7, "时间筛查天数，配合 --time-filter 使用")
	rootCmd.Flags().BoolVar(&enableSuspicious, "suspicious", false, "启用可疑文件扫描，查找压缩包和可执行文件等")
	rootCmd.Flags().StringSliceVar(&suspiciousExtsList, "suspicious-exts", nil, "指定可疑文件扩展名列表，如 .zip,.exe,.rar")
}

func run(cmd *cobra.Command, args []string) {
	// 检查是否没有提供任何参数和标志，如果是则显示帮助
	if len(args) == 0 && !hasAnyFlags() {
		showLogo()
		cmd.Help()
		return
	}

	// 显示Logo
	showLogo()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 设置日志级别
	logger.SetLevel(cfg.LogLevel)

	// 自动设置日志文件
	logFileName := logger.GenerateLogFileName()
	if err := logger.SetLogFile(logFileName); err != nil {
		fmt.Printf("设置日志文件失败: %v\n", err)
	} else {
		defer logger.CloseLogFile()
	}

	logger.Info("SnowFinder 启动，配置加载成功")
	logger.Info("日志文件: %s", logFileName)

	// 创建编码器注册表
	registry := encoder.NewRegistry()
	registry.RegisterDefaultEncoders()

	// 如果要显示编码器，直接显示并退出
	if showEncoders {
		showEncodersList(registry)
		return
	}

	// 处理配置管理命令
	if len(addPatterns) > 0 {
		cfg.AddPatterns(addPatterns)
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("添加匹配词: %v\n", addPatterns)
		fmt.Printf("当前匹配词: %v\n", cfg.Patterns)
		return
	}

	if len(deletePatterns) > 0 {
		cfg.RemovePatterns(deletePatterns)
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("删除匹配词: %v\n", deletePatterns)
		fmt.Printf("当前匹配词: %v\n", cfg.Patterns)
		return
	}

	if len(addKeywords) > 0 {
		cfg.AddMatchKeywords(addKeywords)
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("添加匹配关键字: %v\n", addKeywords)
		fmt.Printf("当前匹配关键字: %v\n", cfg.MatchKeywords)
		return
	}

	if len(deleteKeywords) > 0 {
		cfg.RemoveMatchKeywords(deleteKeywords)
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("删除匹配关键字: %v\n", deleteKeywords)
		fmt.Printf("当前匹配关键字: %v\n", cfg.MatchKeywords)
		return
	}

	// 检查是否提供了搜索路径，如果没有则使用交互模式或默认路径
	var searchPath string
	if len(args) == 0 {
		// 没有提供路径参数
		shouldUseInteractive := (cfg.Interactive && !nonInteractive) || interactiveMode

		if shouldUseInteractive {
			logger.Info("启动交互模式选择搜索路径")
			selector := interactive.NewDirectorySelector()
			var err error
			searchPath, err = selector.SelectPath(cfg.DefaultPath)
			if err != nil {
				logger.Error("交互选择路径失败: %v", err)
				fmt.Printf("交互选择路径失败: %v\n", err)
				os.Exit(1)
			}
		} else {
			// 使用默认路径
			searchPath = cfg.DefaultPath
			logger.Info("使用默认搜索路径: %s", searchPath)
			fmt.Printf("使用默认搜索路径: %s\n", searchPath)
		}
	} else {
		// 使用提供的路径参数
		searchPath = args[0]
		logger.Info("使用命令行指定路径: %s", searchPath)
	}

	// 准备搜索选项
	options := searcher.SearchOptions{
		OutputFile:           outputFile,
		UseKeywordMode:       useKeywordMode,
		ExcludeFiles:         cfg.ExcludeFiles,
		ExcludeExts:          cfg.ExcludeExts,
		BufferSize:           cfg.BufferSize,
		EnableTimeFilter:     enableTimeFilter || cfg.EnableTimeFilter,
		TimeFilterDays:       timeFilterDays,
		EnableSuspiciousScan: enableSuspicious || cfg.EnableSuspiciousScan,
		SuspiciousExts:       cfg.SuspiciousExts,
	}

	// 如果命令行指定了时间筛查天数，使用命令行的值，否则使用配置文件的值
	if timeFilterDays <= 0 {
		options.TimeFilterDays = cfg.TimeFilterDays
	}

	// 如果命令行指定了可疑扩展名，使用命令行的值
	if len(suspiciousExtsList) > 0 {
		options.SuspiciousExts = suspiciousExtsList
	}

	// 确定使用的模式和关键词
	hasKeywordSearch := false
	if len(patterns) > 0 {
		// 使用命令行指定的模式
		options.Patterns = patterns
		fmt.Printf("匹配词: %v\n", patterns)
		hasKeywordSearch = true
	} else if useKeywordMode {
		// 使用关键词模式
		options.MatchKeywords = cfg.MatchKeywords
		fmt.Printf("匹配关键字: %v\n", cfg.MatchKeywords)
		hasKeywordSearch = true
	} else if !options.EnableSuspiciousScan {
		// 只有在没有启用可疑文件扫描时，才要求必须有关键词搜索
		options.Patterns = cfg.Patterns
		if len(options.Patterns) == 0 {
			fmt.Println("警告：没有指定匹配词且未启用可疑文件扫描。请使用 --add 添加匹配词、使用 -p 指定匹配词，或使用 --suspicious 启用可疑文件扫描。")
			return
		}
		fmt.Printf("匹配词: %v\n", options.Patterns)
		hasKeywordSearch = true
	} else {
		// 启用了可疑文件扫描，但没有关键词搜索
		fmt.Println("仅启用可疑文件扫描模式")
	}

	// 显示启用的功能
	var enabledFeatures []string
	if hasKeywordSearch {
		if useKeywordMode {
			enabledFeatures = append(enabledFeatures, "关键词搜索")
		} else {
			enabledFeatures = append(enabledFeatures, "编码匹配搜索")
		}
	}
	if options.EnableSuspiciousScan {
		enabledFeatures = append(enabledFeatures, "可疑文件检测")
	}
	if options.EnableTimeFilter {
		enabledFeatures = append(enabledFeatures, fmt.Sprintf("时间筛查(%d天)", options.TimeFilterDays))
	}

	if len(enabledFeatures) > 0 {
		fmt.Printf("启用功能: %s\n", strings.Join(enabledFeatures, ", "))
	}

	// 在交互模式下显示搜索确认
	if (cfg.Interactive && !nonInteractive) || interactiveMode {
		selector := interactive.NewDirectorySelector()
		if !selector.ConfirmSearch(searchPath, options.Patterns, options.MatchKeywords, useKeywordMode) {
			fmt.Println("用户取消搜索")
			return
		}
	}

	fmt.Println() // 空行分隔

	// 执行搜索
	startTime := time.Now()

	// 使用配置中的工作线程数
	if workers <= 0 {
		workers = cfg.MaxWorkers
	}

	searchEngine := searcher.NewSearcher(registry, workers, cfg.ExcludeFiles, cfg.ExcludeExts, cfg.BufferSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results, err := searchEngine.Search(ctx, searchPath, options)
	if err != nil {
		logger.Error("搜索失败: %v", err)
		fmt.Printf("搜索失败: %v\n", err)
		os.Exit(1)
	}

	// 显示结果
	for _, result := range results {
		fmt.Print(result.String())
		fmt.Println()
	}

	elapsed := time.Since(startTime)

	if useKeywordMode {
		fmt.Printf("以上结果基于匹配关键字: %v\n", cfg.MatchKeywords)
	} else {
		fmt.Printf("以上结果基于匹配词: %v, 添加更多匹配词请使用 --add 参数\n", options.Patterns)
	}
	fmt.Printf("搜索完成，用时：%.2f秒\n", elapsed.Seconds())

	if len(results) > 0 {
		color.Green("找到 %d 个匹配结果", len(results))
	} else {
		color.Yellow("未找到匹配结果")
	}
}

func showLogo() {
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	logo := `
       ______    _                __              
      / ____/   (_)   ____   ____/ /  ___    _____
     / /_      / /   / __ \ / __  /  / _ \  / ___/
    / __/     / /   / / / // /_/ /  /  __/ / /    
   /_/       /_/   /_/ /_/ \__,_/   \___/ /_/     
`

	cyan.Print(logo)
	yellow.Println(strings.Repeat(" ", 70) + "- V 2.1.0 (Go)")
	yellow.Println("Usage: snowfind [path] [-i] [-o output] [-p pattern] [--add add] [--del delete] [--show-encoders] [--match] [--time-filter] [--suspicious]")
	yellow.Println("Example: snowfind hack.pcap -o result.txt")
	yellow.Println("         snowfind -i  (交互模式)")
	yellow.Println("         snowfind --time-filter --days 3 --suspicious  (检测3天内的可疑文件)")
	fmt.Println()
}

func showEncodersList(registry *encoder.Registry) {
	fmt.Println("已加载的编码器：")
	encoderList := registry.ListEncoders()
	for _, info := range encoderList {
		fmt.Printf("- %s\n", info)
	}
	fmt.Println("\n您可以自定义自己需要的编码器，编码器需要实现Encoder接口。")
}

// hasAnyFlags 检查是否设置了任何标志
func hasAnyFlags() bool {
	return outputFile != "" ||
		len(patterns) > 0 ||
		len(addPatterns) > 0 ||
		len(deletePatterns) > 0 ||
		showEncoders ||
		useKeywordMode ||
		len(addKeywords) > 0 ||
		len(deleteKeywords) > 0 ||
		workers != 4 || // 默认值是4，如果不等于4说明被设置了
		interactiveMode ||
		nonInteractive ||
		enableTimeFilter ||
		timeFilterDays != 7 || // 默认值是7，如果不等于7说明被设置了
		enableSuspicious ||
		len(suspiciousExtsList) > 0
}
