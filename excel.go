package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

func generateExcel(req ExcelRequest, writer io.Writer) error {
	// Set defaults
	req.SetDefaults()
	
	// Validate input
	if len(req.Headers) == 0 {
		return fmt.Errorf("headers cannot be empty")
	}
	
	if len(req.Data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}
	
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()
	
	// Create or rename the default sheet
	sheetName := req.SheetName
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	
	// Rename default sheet if needed
	if sheetName != "Sheet1" {
		f.SetSheetName("Sheet1", sheetName)
	}
	
	// Set headers
	for i, header := range req.Headers {
		cell := fmt.Sprintf("%s1", columnIndexToLetter(i))
		f.SetCellValue(sheetName, cell, header)
	}
	
	// Apply header styling if provided
	if req.Styles != nil && req.Styles.HeaderStyle != nil {
		headerStyle, err := createExcelStyle(f, req.Styles.HeaderStyle)
		if err != nil {
			return fmt.Errorf("failed to create header style: %v", err)
		}
		
		// Apply header style to all header cells
		for i := range req.Headers {
			cell := fmt.Sprintf("%s1", columnIndexToLetter(i))
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}
	
	// Set data
	for rowIndex, row := range req.Data {
		excelRow := rowIndex + 2 // Start from row 2 (after headers)
		
		// Ensure row length matches headers
		if len(row) != len(req.Headers) {
			return fmt.Errorf("row %d has %d columns, expected %d", rowIndex+1, len(row), len(req.Headers))
		}
		
		for colIndex, cell := range row {
			cellAddress := fmt.Sprintf("%s%d", columnIndexToLetter(colIndex), excelRow)
			f.SetCellValue(sheetName, cellAddress, cell)
		}
	}
	
	// Apply data styling if provided
	if req.Styles != nil && req.Styles.DataStyle != nil {
		dataStyle, err := createExcelStyle(f, req.Styles.DataStyle)
		if err != nil {
			return fmt.Errorf("failed to create data style: %v", err)
		}
		
		// Apply data style to all data cells
		if len(req.Data) > 0 {
			startCell := "A2"
			endCell := fmt.Sprintf("%s%d", columnIndexToLetter(len(req.Headers)-1), len(req.Data)+1)
			f.SetCellStyle(sheetName, startCell, endCell, dataStyle)
		}
	}
	
	// Auto-size columns if requested
	if req.AutoSize {
		for i := range req.Headers {
			col := columnIndexToLetter(i)
			f.SetColWidth(sheetName, col, col, 15) // Set a reasonable default width
		}
	}
	
	// Write to the provided writer
	return f.Write(writer)
}

func createExcelStyle(f *excelize.File, style *CellStyle) (int, error) {
	styleConfig := &excelize.Style{}
	
	// Font configuration
	font := &excelize.Font{}
	if style.Bold {
		font.Bold = true
	}
	if style.FontSize > 0 {
		font.Size = float64(style.FontSize)
	}
	if style.FontColor != "" {
		font.Color = strings.TrimPrefix(style.FontColor, "#")
	}
	styleConfig.Font = font
	
	// Background color
	if style.Background != "" {
		fill := &excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{strings.TrimPrefix(style.Background, "#")},
		}
		styleConfig.Fill = *fill
	}
	
	// Alignment
	if style.Alignment != "" {
		alignment := &excelize.Alignment{}
		switch strings.ToLower(style.Alignment) {
		case "left":
			alignment.Horizontal = "left"
		case "center":
			alignment.Horizontal = "center"
		case "right":
			alignment.Horizontal = "right"
		}
		styleConfig.Alignment = alignment
	}
	
	return f.NewStyle(styleConfig)
}

func columnIndexToLetter(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

func sanitizeExcelFilename(filename string) string {
	// Remove path separators and dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	
	// Ensure it has .xlsx extension
	if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		filename += ".xlsx"
	}
	
	return filename
}

func sanitizeSheetName(name string) string {
	// Excel sheet name restrictions
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "[", "_")
	name = strings.ReplaceAll(name, "]", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, ":", "_")
	
	// Limit length to 31 characters
	if len(name) > 31 {
		name = name[:31]
	}
	
	return name
}