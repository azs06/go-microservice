package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func generateCSV(req CSVRequest, writer io.Writer) error {
	// Set defaults
	req.SetDefaults()
	
	// Validate input
	if len(req.Headers) == 0 {
		return fmt.Errorf("headers cannot be empty")
	}
	
	if len(req.Data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}
	
	// Create CSV writer
	csvWriter := csv.NewWriter(writer)
	if req.Delimiter != "," && len(req.Delimiter) == 1 {
		csvWriter.Comma = rune(req.Delimiter[0])
	}
	
	// Write headers
	if err := csvWriter.Write(req.Headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}
	
	// Write data rows
	for i, row := range req.Data {
		// Convert row to strings
		stringRow := make([]string, len(row))
		for j, cell := range row {
			stringRow[j] = convertToString(cell)
		}
		
		// Ensure row length matches headers
		if len(stringRow) != len(req.Headers) {
			return fmt.Errorf("row %d has %d columns, expected %d", i+1, len(stringRow), len(req.Headers))
		}
		
		if err := csvWriter.Write(stringRow); err != nil {
			return fmt.Errorf("failed to write row %d: %v", i+1, err)
		}
	}
	
	csvWriter.Flush()
	return csvWriter.Error()
}

func convertToString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func sanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	
	// Ensure it has .csv extension
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		filename += ".csv"
	}
	
	return filename
}