package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/parser/fs4"
	"github.com/Civil/mlx5fw-go/pkg/section"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func parseReplaceSectionArgs(args []string) (string, int, error) {
	if len(args) < 1 {
		return "", -1, merry.New("section name required")
	}

	sectionArg := args[0]
	parts := strings.Split(sectionArg, ":")
	sectionName := parts[0]
	sectionID := -1

	if len(parts) > 1 {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", -1, merry.Wrap(err)
		}
		sectionID = id
	}

	return sectionName, sectionID, nil
}

func runReplaceSectionCommand(cmd *cobra.Command, args []string, sectionName string, sectionID int, replacementFile string, outputFile string) error {
	logger.Debug("Starting replace-section command",
		zap.String("firmware", firmwarePath),
		zap.String("section", sectionName),
		zap.Int("id", sectionID),
		zap.String("replacement", replacementFile),
		zap.String("output", outputFile))

	// Read the entire firmware file
	firmwareData, err := os.ReadFile(firmwarePath)
	if err != nil {
		return merry.Wrap(err)
	}

	// Read the replacement data
	replacementData, err := os.ReadFile(replacementFile)
	if err != nil {
		return merry.Wrap(err)
	}

	// Parse the firmware using the existing parser
	reader, err := parser.NewFirmwareReader(firmwarePath, logger)
	if err != nil {
		return merry.Wrap(err)
	}
	defer reader.Close()

	fwParser := fs4.NewParser(reader, logger)
	err = fwParser.Parse()
	if err != nil {
		return merry.Wrap(err)
	}

	// Find the target section
	allSections := fwParser.GetSections()
	var targetSection interfaces.SectionInterface

	for _, sections := range allSections {
		for idx, section := range sections {
			sectionTypeName := types.GetSectionTypeName(section.Type())
			if sectionTypeName == sectionName && (sectionID == -1 || idx == sectionID) {
				targetSection = section
				break
			}
		}
		if targetSection != nil {
			break
		}
	}

	if targetSection == nil {
		return merry.New(fmt.Sprintf("section '%s' with ID %d not found", sectionName, sectionID))
	}

	logger.Info("Found target section",
		zap.String("name", sectionName),
		zap.Uint64("offset", targetSection.Offset()),
		zap.Uint32("size", targetSection.Size()),
		zap.Uint32("newSize", uint32(len(replacementData))))

	// Create section replacer with parser context (use fully-featured implementation)
	replacer := section.NewReplacer(fwParser, firmwareData, logger)

	// Replace the section
	newFirmwareData, err := replacer.ReplaceSection(targetSection, replacementData)
	if err != nil {
		return merry.Wrap(err)
	}

	// Write the modified firmware
	err = os.WriteFile(outputFile, newFirmwareData, 0644)
	if err != nil {
		return merry.Wrap(err)
	}

	logger.Info("Successfully replaced section and wrote output file",
		zap.String("output", outputFile))

	return nil
}
