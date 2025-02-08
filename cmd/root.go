package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cf-downloader/core/config"
	"cf-downloader/core/request"

	"github.com/spf13/cobra"
)

var (
	cfg = config.New()
)

var rootCmd = &cobra.Command{
	Use:   "cf-downloader",
	Short: "A CF downloader",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 确保输出目录存在
		if cfg.OutputDir == "" {

			cfg.OutputDir = "downloads"
		}

		// 转换为绝对路径
		absPath, err := filepath.Abs(cfg.OutputDir)
		if err != nil {
			fmt.Printf("Error resolving output path: %v\n", err)
			os.Exit(1)
		}
		cfg.OutputDir = absPath

		// 创建输出目录
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Create a new client with the configured proxy and retry settings
		client := request.NewClient(cfg.ProxyURL, cfg.MaxRetries, cfg.RetryDelay)

		// Parse custom headers
		var requestOpts *request.RequestOption
		if len(cfg.Headers) > 0 {
			headers := make(map[string]string)
			for _, header := range cfg.Headers {
				parts := strings.SplitN(header, "=", 2)
				if len(parts) == 2 {
					headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
			if len(headers) > 0 {
				requestOpts = &request.RequestOption{
					Headers: headers,
				}
			}
		}

		// Generate output filename from the URL
		filename := filepath.Base(cfg.URL)
		if filename == "" || filename == "." {
			filename = "downloaded_file"
		}
		outputPath := filepath.Join(cfg.OutputDir, filename)

		fmt.Printf("Downloading %s to %s\n", cfg.URL, outputPath)

		// Download the file with custom headers
		if err := client.DownloadFile(cfg.URL, outputPath, requestOpts); err != nil {
			fmt.Printf("Download error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Download completed successfully!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 必需参数
	rootCmd.Flags().StringVarP(&cfg.URL, "url", "u", "", "Target URL to crawl")
	rootCmd.MarkFlagRequired("url")

	// 可选参数
	rootCmd.Flags().StringVarP(&cfg.ProxyURL, "proxy", "p", "", "Proxy URL (optional)")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 5, "Number of concurrent downloads")
	rootCmd.Flags().StringVarP(&cfg.OutputDir, "output", "o", "downloads", "Output directory for downloaded files")

	// Add new flags for retry configuration
	rootCmd.Flags().IntVarP(&cfg.MaxRetries, "max-retries", "r", 3, "Maximum number of retry attempts for downloads")
	rootCmd.Flags().IntVarP(&cfg.RetryDelay, "retry-delay", "d", 5, "Delay between retry attempts in seconds")

	// Add flag for custom headers
	rootCmd.Flags().StringArrayVarP(&cfg.Headers, "header", "H", nil,
		"Custom headers in key=value format (can be used multiple times)")
}
