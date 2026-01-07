package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"ai-manager/internal/cleanup"
	"ai-manager/internal/config"
	"ai-manager/internal/discovery"
	"ai-manager/internal/models"
	"ai-manager/internal/utils"

	"github.com/spf13/cobra"
)

var (
	days int
	verbose bool
	jsonOutput bool
)

// newScanCmd returns the scan command with implementation
func newScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for AI tools on your system",
		Long: `Scan and discover AI tools installed on your system.
Shows which tools are found, their paths, and disk usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			scanner := discovery.NewScanner(cfg)
			result, err := scanner.Scan()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(result)
			}

			printScanResult(result, verbose)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newCleanupCmd returns the cleanup command with implementation
func newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up temporary files",
		Long: `Clean up temporary files from AI tools.
By default, removes files older than 7 days.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			// Override retention days if specified
			if days > 0 {
				cfg.Retention.TempFiles = days
			}

			cleaner := cleanup.NewCleaner(cfg)
			results, err := cleaner.CleanupAll()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(results)
			}

			totalFreed := int64(0)
			totalDeleted := 0

			fmt.Println("=== Cleanup Results ===")
			for _, r := range results {
				fmt.Println(cleanup.FormatResult(r))
				totalFreed += r.SpaceFreed
				totalDeleted += r.FilesDeleted
			}

			fmt.Printf("\nTotal: %d files deleted, %s freed\n",
				totalDeleted, models.FormatBytes(totalFreed))

			return nil
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 7, "Delete files older than N days")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newCheckCmd returns the health check command
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Health check for AI tools",
		Long: `Run health checks on your AI tools and configurations.
Reports on configuration validity, broken links, and disk usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			fmt.Println("=== AI Tools Health Check ===\n")

			issues := 0
			for _, tool := range cfg.Tools {
				if !tool.Enabled {
					continue
				}

				toolPath := utils.ExpandPath(tool.Path)
				configPath := utils.ExpandPath(tool.ConfigPath)

				fmt.Printf("[%s]\n", tool.Name)

				// Check if tool exists
				if _, err := os.Stat(toolPath); os.IsNotExist(err) {
					fmt.Printf("  ✗ Tool path not found: %s\n", toolPath)
					issues++
					continue
				}
				fmt.Printf("  ✓ Path exists: %s\n", toolPath)

				// Check config file
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					fmt.Printf("  ⚠ Config file missing: %s\n", configPath)
					issues++
				} else {
					fmt.Printf("  ✓ Config file found: %s\n", configPath)
				}

				// Check symlinks if any
				fmt.Println()
			}

			if issues > 0 {
				fmt.Printf("Found %d issue(s)\n", issues)
			} else {
				fmt.Println("All tools are healthy!")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newStatsCmd returns the stats command
func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show usage statistics",
		Long: `Show usage statistics and disk usage for AI tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			fmt.Println("=== AI Tools Disk Usage ===\n")

			totalSize := int64(0)
			totalFiles := 0

			for _, tool := range cfg.Tools {
				if !tool.Enabled {
					continue
				}

				path := utils.ExpandPath(tool.Path)
				size, _ := utils.FileSize(path)
				files, _ := utils.CountFiles(path)

				totalSize += size
				totalFiles += files

				fmt.Printf("[%s]\n", tool.Name)
				fmt.Printf("  Path: %s\n", path)
				fmt.Printf("  Size: %s (%d files)\n\n", utils.FormatSize(size), files)
			}

			fmt.Printf("Total: %s (%d files across %d tools)\n",
				utils.FormatSize(totalSize), totalFiles, len(cfg.Tools))

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// Helper functions
func printScanResult(result *models.ScanResult, verbose bool) {
	fmt.Printf("=== Scan Results (%d tools found, %d enabled) ===\n\n",
		result.Total, result.Enabled)

	for _, tool := range result.Tools {
		status := "✓"
		if tool.Status == models.StatusNotFound {
			status = "✗"
		} else if tool.Status == models.StatusWarning {
			status = "⚠"
		}

		fmt.Printf("%s [%s]\n", status, tool.Name)

		if tool.Found {
			fmt.Printf("  Path: %s\n", tool.Path)
			if verbose {
				fmt.Printf("  Config: %s\n", tool.ConfigPath)
				fmt.Printf("  Data: %s\n", tool.DataPath)
				fmt.Printf("  Disk: %s (%d files)\n",
					models.FormatBytes(tool.DiskUsage.SizeBytes),
					tool.DiskUsage.Files)
			}
		} else {
			fmt.Printf("  Not found on system\n")
		}
		fmt.Println()
	}
}

// printJSON outputs data as JSON
func printJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
