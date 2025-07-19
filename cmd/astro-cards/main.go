package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nilspolek/AstroCards/internal/cardgen"
	"github.com/nilspolek/AstroCards/internal/pdfreader"
)

func main() {
	// Command-line flags
	pdfPath := flag.String("pdf", "", "Path to the PDF file")
	ollamaURL := flag.String("ollama", "http://localhost:11434", "Ollama server URL")
	model := flag.String("model", "llama3.2", "Ollama model name")
	outputPath := flag.String("output", "", "Output file for the generated cards (default: stdout)")
	flag.Parse()

	if *pdfPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --pdf flag is required")
		os.Exit(1)
	}

	// Read PDF pages
	pages, err := pdfreader.ReadPDF(*pdfPath)
	if err != nil {
		log.Fatalf("Failed to read PDF: %v", err)
	}

	// Create card generator
	cg := cardgen.NewCardGenerator(*ollamaURL, *model)

	// Generate cards asynchronously with progress
	progress := make(chan int)
	resultChan, errChan := cg.GenerateCardsAsync(pages, progress)

	// Print progress to stderr as a progress bar
	const barWidth = 40
	go func() {
		for p := range progress {
			filled := int(float64(barWidth) * float64(p) / 100.0)
			bar := "[" + string(repeatRune('=', filled)) + ">" + string(repeatRune(' ', barWidth-filled)) + "]"
			fmt.Fprintf(os.Stderr, "\rProgress: %3d%% %s", p, bar)
		}
		fmt.Fprintln(os.Stderr)
	}()

	cards := <-resultChan
	if err := <-errChan; err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred while generating: %s\n", err.Error())
	}

	// Output cards as JSON
	var out *os.File
	if *outputPath != "" {
		var err error
		out, err = os.Create(*outputPath)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cards); err != nil {
		log.Fatalf("Failed to encode cards as JSON: %v", err)
	}
}

// Helper function to repeat a rune n times and return a string
func repeatRune(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	runes := make([]rune, n)
	for i := range runes {
		runes[i] = r
	}
	return string(runes)
}
