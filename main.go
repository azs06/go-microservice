// main.go
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check API key if configured
		if apiKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("PDF microservice starting on :%s", *port)
	if apiKey != "" {
		log.Println("API key authentication enabled")
	} else {
		log.Println("Running in open access mode (no API key required)")
	}
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

