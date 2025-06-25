package main

// Request structures
type CSVRequest struct {
	Data      [][]interface{} `json:"data"`
	Headers   []string        `json:"headers"`
	Filename  string          `json:"filename,omitempty"`
	Delimiter string          `json:"delimiter,omitempty"`
	Encoding  string          `json:"encoding,omitempty"`
}

type ExcelRequest struct {
	Data      [][]interface{} `json:"data"`
	Headers   []string        `json:"headers"`
	Filename  string          `json:"filename,omitempty"`
	SheetName string          `json:"sheet_name,omitempty"`
	AutoSize  bool            `json:"auto_size,omitempty"`
	Styles    *ExcelStyles    `json:"styles,omitempty"`
}

type ExcelStyles struct {
	HeaderStyle *CellStyle `json:"header_style,omitempty"`
	DataStyle   *CellStyle `json:"data_style,omitempty"`
}

type CellStyle struct {
	Bold       bool   `json:"bold,omitempty"`
	FontSize   int    `json:"font_size,omitempty"`
	FontColor  string `json:"font_color,omitempty"`
	Background string `json:"background,omitempty"`
	Alignment  string `json:"alignment,omitempty"`
}

// Response structures
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
	Version  string            `json:"version"`
}

// Helper methods
func (c *CSVRequest) SetDefaults() {
	if c.Delimiter == "" {
		c.Delimiter = ","
	}
	if c.Encoding == "" {
		c.Encoding = "UTF-8"
	}
	if c.Filename == "" {
		c.Filename = "export.csv"
	}
}

func (e *ExcelRequest) SetDefaults() {
	if e.SheetName == "" {
		e.SheetName = "Sheet1"
	}
	if e.Filename == "" {
		e.Filename = "export.xlsx"
	}
}

