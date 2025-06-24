// main.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/page"
)

var (
	apiKey = os.Getenv("API_KEY") // Optional API key from environment
	port   = flag.String("port", "8080", "Port to run the server on")
)

func main() {
	flag.Parse()
	
	http.HandleFunc("/pdf", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check API key if configured
		if apiKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		html := r.FormValue("html")
		if html == "" {
			http.Error(w, "Missing html parameter", http.StatusBadRequest)
			return
		}

		// Create context for this request
		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		var pdfBuf []byte

		err := chromedp.Run(ctx,
			chromedp.Navigate("data:text/html,"+url.QueryEscape(html)),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				pdfBuf, _, err = page.PrintToPDF().
					WithPrintBackground(true).
					WithPaperWidth(8.27).  // A4 width in inches
					WithPaperHeight(11.7). // A4 height in inches
					WithMarginTop(0.4).
					WithMarginBottom(0.4).
					WithMarginLeft(0.4).
					WithMarginRight(0.4).
					Do(ctx)
				return err
			}),
		)
		if err != nil {
			log.Printf("PDF generation error: %v", err)
			http.Error(w, "PDF generation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", string(rune(len(pdfBuf))))
		w.Write(pdfBuf)
	})

	// CSV generation endpoint
	http.HandleFunc("/csv", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check API key if configured
		if apiKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey != apiKey {
				writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized, "")
				return
			}
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			writeErrorResponse(w, "Failed to parse form data", http.StatusBadRequest, err.Error())
			return
		}

		// Build CSV request
		var csvReq CSVRequest
		
		// Parse data (required)
		if err := parseFormData(r.FormValue("data"), &csvReq.Data); err != nil {
			writeErrorResponse(w, "Invalid data format", http.StatusBadRequest, err.Error())
			return
		}
		
		// Parse headers (required)
		if err := parseFormData(r.FormValue("headers"), &csvReq.Headers); err != nil {
			writeErrorResponse(w, "Invalid headers format", http.StatusBadRequest, err.Error())
			return
		}
		
		// Optional parameters
		csvReq.Filename = r.FormValue("filename")
		csvReq.Delimiter = r.FormValue("delimiter")
		csvReq.Encoding = r.FormValue("encoding")
		
		// Set defaults
		csvReq.SetDefaults()
		
		// Validate input
		if len(csvReq.Headers) == 0 {
			writeErrorResponse(w, "Headers cannot be empty", http.StatusBadRequest, "")
			return
		}
		
		if len(csvReq.Data) == 0 {
			writeErrorResponse(w, "Data cannot be empty", http.StatusBadRequest, "")
			return
		}

		// Generate CSV
		var buf bytes.Buffer
		if err := generateCSV(csvReq, &buf); err != nil {
			log.Printf("CSV generation error: %v", err)
			writeErrorResponse(w, "CSV generation failed", http.StatusInternalServerError, err.Error())
			return
		}

		// Set response headers
		filename := sanitizeFilename(csvReq.Filename)
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		
		// Write response
		w.Write(buf.Bytes())
	})

	// Excel generation endpoint
	http.HandleFunc("/excel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check API key if configured
		if apiKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey != apiKey {
				writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized, "")
				return
			}
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			writeErrorResponse(w, "Failed to parse form data", http.StatusBadRequest, err.Error())
			return
		}

		// Build Excel request
		var excelReq ExcelRequest
		
		// Parse data (required)
		if err := parseFormData(r.FormValue("data"), &excelReq.Data); err != nil {
			writeErrorResponse(w, "Invalid data format", http.StatusBadRequest, err.Error())
			return
		}
		
		// Parse headers (required)
		if err := parseFormData(r.FormValue("headers"), &excelReq.Headers); err != nil {
			writeErrorResponse(w, "Invalid headers format", http.StatusBadRequest, err.Error())
			return
		}
		
		// Optional parameters
		excelReq.Filename = r.FormValue("filename")
		excelReq.SheetName = r.FormValue("sheet_name")
		
		// Parse auto_size
		if autoSizeStr := r.FormValue("auto_size"); autoSizeStr != "" {
			if autoSize, err := strconv.ParseBool(autoSizeStr); err == nil {
				excelReq.AutoSize = autoSize
			}
		} else {
			excelReq.AutoSize = true // default
		}
		
		// Parse styles if provided
		if stylesStr := r.FormValue("styles"); stylesStr != "" {
			if err := parseFormData(stylesStr, &excelReq.Styles); err != nil {
				writeErrorResponse(w, "Invalid styles format", http.StatusBadRequest, err.Error())
				return
			}
		}
		
		// Set defaults
		excelReq.SetDefaults()
		
		// Validate input
		if len(excelReq.Headers) == 0 {
			writeErrorResponse(w, "Headers cannot be empty", http.StatusBadRequest, "")
			return
		}
		
		if len(excelReq.Data) == 0 {
			writeErrorResponse(w, "Data cannot be empty", http.StatusBadRequest, "")
			return
		}

		// Generate Excel
		var buf bytes.Buffer
		if err := generateExcel(excelReq, &buf); err != nil {
			log.Printf("Excel generation error: %v", err)
			writeErrorResponse(w, "Excel generation failed", http.StatusInternalServerError, err.Error())
			return
		}

		// Set response headers
		filename := sanitizeExcelFilename(excelReq.Filename)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		
		// Write response
		w.Write(buf.Bytes())
	})

	// Enhanced health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check API key if configured
		if apiKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey != apiKey {
				writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized, "")
				return
			}
		}
		
		// Enhanced health response
		health := HealthResponse{
			Status: "ok",
			Services: map[string]string{
				"pdf":   "available",
				"csv":   "available",
				"excel": "available",
			},
			Version: "1.0.0",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})

	log.Printf("Document generation microservice starting on :%s", *port)
	log.Println("Available endpoints: /pdf, /csv, /excel, /health")
	if apiKey != "" {
		log.Println("API key authentication enabled")
	} else {
		log.Println("Running in open access mode (no API key required)")
	}
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

