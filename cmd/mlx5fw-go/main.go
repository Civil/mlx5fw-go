package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	// Logger for the application
	logger *zap.Logger

	// Command line flags
	verboseLogging bool
	firmwarePath   string
	jsonOutput     bool
)

func main() {
	// Parse command line args early to check for JSON flag
	// We need to do this before logger initialization
	for _, arg := range os.Args[1:] {
		if arg == "--json" {
			jsonOutput = true
			break
		}
	}

	// Initialize zap logger
	var err error
	logger, err = ConfigureLogger(jsonOutput, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	rootCmd := &cobra.Command{
		Use:   "mlx5fw-go",
		Short: "Mellanox firmware parsing tool",
		Long: `A CLI tool for parsing and analyzing Mellanox firmware files (FS4 and FS5 formats)

This tool provides functionality to analyze Mellanox firmware files including:
- Listing firmware sections
- Displaying section contents in human-readable format
- Parsing ITOC (Image Table of Contents) and DTOC (Device Table of Contents)

Example usage:
  mlx5fw-go sections -f firmware.bin                # List all sections
  mlx5fw-go sections -f firmware.bin -c             # Show section contents
  mlx5fw-go sections -f firmware.bin -v             # Enable verbose logging`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verboseLogging, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&firmwarePath, "file", "f", "", "Firmware file path (required)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	// Setup verbose logging after flags are parsed
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// If verbose logging is requested, reconfigure the logger
		if verboseLogging {
			logger, err = ConfigureLogger(jsonOutput, true)
			if err != nil {
				return fmt.Errorf("failed to initialize verbose logger: %w", err)
			}
		}
		
		// Add firmware filename to logger context if provided
		if firmwarePath != "" {
			logger = logger.With(zap.String("firmware", firmwarePath))
		}
		
		return nil
	}


	// Add sections command
	sectionsCmd := &cobra.Command{
		Use:   "sections",
		Short: "Print firmware sections",
		Long:  "Display all sections found in the firmware file in human-readable format",
	}
	
	// Add flags
	var showContent bool
	sectionsCmd.Flags().BoolVarP(&showContent, "content", "c", false, "Show section content")
	
	// Store the flag value for use in command
	sectionsCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := ValidateFirmwarePath(firmwarePath); err != nil {
			return err
		}
		outputFormat := ""
		if jsonOutput {
			outputFormat = "json"
		}
		return runSectionsCommand(cmd, args, showContent, outputFormat)
	}

	rootCmd.AddCommand(sectionsCmd)

	// Add query command
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Query firmware metadata",
		Long:  "Display firmware metadata information similar to mstflint query output",
	}
	
	// Add flags
	var fullOutput bool
	queryCmd.Flags().BoolVar(&fullOutput, "full", false, "Show full query output")
	
	// Store the flag value for use in command
	queryCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := ValidateFirmwarePath(firmwarePath); err != nil {
			return err
		}
		return runQueryCommand(cmd, args, fullOutput, jsonOutput)
	}

	rootCmd.AddCommand(queryCmd)

	// Add print-config command
	printConfigCmd := &cobra.Command{
		Use:   "print-config",
		Short: "Print firmware configuration (INI)",
		Long:  "Display the DBG_FW_INI section from the firmware file (equivalent to mstflint dc command)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ValidateFirmwarePath(firmwarePath); err != nil {
				return err
			}
			return runPrintConfigCommand(cmd, args)
		},
	}
	rootCmd.AddCommand(printConfigCmd)

	// Add extract command
	extractCmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract firmware sections to files",
		Long:  "Extract all firmware sections as separate files to the specified output directory (CRC is always removed and metadata is always included)",
	}
	
	// Add output directory flag
	var extractOutputDir string
	extractCmd.Flags().StringVarP(&extractOutputDir, "output", "o", ".", "Output directory for extracted sections")
	
	// Add JSON export flag (deprecated, kept for backwards compatibility)
	extractCmd.Flags().Bool("json", false, "DEPRECATED: JSON is now always exported for parsed sections")
	
	// Add keep-binary flag
	extractCmd.Flags().Bool("keep-binary", false, "Keep binary representation alongside JSON (by default, only JSON is saved for parsed sections)")
	
	extractCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := ValidateFirmwarePath(firmwarePath); err != nil {
			return err
		}
		return runExtractCommand(cmd, args, extractOutputDir)
	}
	
	rootCmd.AddCommand(extractCmd)

	// Add replace-section command
	replaceSectionCmd := &cobra.Command{
		Use:   "replace-section SECTION_NAME[:ID] -r REPLACEMENT_FILE -o OUTPUT_FILE",
		Short: "Replace a section with content from a file",
		Long: `Replace a firmware section with content from a file and recalculate checksums.

This command allows you to replace any section in the firmware with new content.
If the section size changes, the firmware structure will be adjusted accordingly.

Examples:
  mlx5fw-go replace-section -f firmware.bin DBG_FW_INI -r new_ini.txt -o modified.bin
  mlx5fw-go replace-section -f firmware.bin ITOC:0 -r new_itoc.bin -o modified.bin`,
		Args: cobra.ExactArgs(1),
	}
	
	// Add flags
	var replacementFile string
	var outputFile string
	replaceSectionCmd.Flags().StringVarP(&replacementFile, "replacement", "r", "", "File containing replacement data (required)")
	replaceSectionCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output firmware file (required)")
	replaceSectionCmd.MarkFlagRequired("replacement")
	replaceSectionCmd.MarkFlagRequired("output")
	
	replaceSectionCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := ValidateFirmwarePath(firmwarePath); err != nil {
			return err
		}
		sectionName, sectionID, err := parseReplaceSectionArgs(args)
		if err != nil {
			return err
		}
		return runReplaceSectionCommandV4(cmd, args, sectionName, sectionID, replacementFile, outputFile)
	}
	
	rootCmd.AddCommand(replaceSectionCmd)

	// Add reassemble command
	reassembleCmd := &cobra.Command{
		Use:   "reassemble",
		Short: "Reassemble firmware from extracted sections",
		Long: `Reassemble a firmware image from previously extracted sections and metadata.

This command reconstructs a complete firmware image from the extracted sections,
gaps, and metadata created by the extract command. It handles CRC reconstruction
and proper section placement according to the original firmware layout.

Examples:
  mlx5fw-go reassemble -i extracted_fw -o reassembled.bin
  mlx5fw-go reassemble -i extracted_fw -o reassembled.bin --verify-crc`,
	}
	
	// Add flags
	var reassembleInputDir string
	var reassembleOutputFile string
	var reassembleVerifyCRC bool
	reassembleCmd.Flags().StringVarP(&reassembleInputDir, "input", "i", "", "Input directory containing extracted sections (required)")
	reassembleCmd.Flags().StringVarP(&reassembleOutputFile, "output", "o", "", "Output firmware file (required)")
	reassembleCmd.Flags().BoolVar(&reassembleVerifyCRC, "verify-crc", false, "Verify CRC values during reassembly")
	reassembleCmd.Flags().Bool("binary-only", false, "Force binary-only mode, ignore JSON files (by default, JSON is preferred)")
	reassembleCmd.MarkFlagRequired("input")
	reassembleCmd.MarkFlagRequired("output")
	
	reassembleCmd.RunE = func(cmd *cobra.Command, args []string) error {
		binaryOnly, _ := cmd.Flags().GetBool("binary-only")
		opts := ReassembleOptions{
			InputDir:   reassembleInputDir,
			OutputFile: reassembleOutputFile,
			VerifyCRC:  reassembleVerifyCRC,
			BinaryOnly: binaryOnly,
		}
		return runReassembleCommand(cmd, args, opts)
	}
	
	rootCmd.AddCommand(reassembleCmd)

	// Add report commands
	rootCmd.AddCommand(CreateReportCommand())
	rootCmd.AddCommand(CreateSectionReportCommand())

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", zap.Error(err))
		os.Exit(1)
	}
}


