package pdfreader

import (
	"fmt"

	"github.com/ledongthuc/pdf"
)

// Page represents a single page in a PDF with its number and text content.
type Page struct {
	Number  int
	Content string
}

// ReadPDF reads the content of a PDF file at the given path and returns a slice of Page structs.
func ReadPDF(path string) ([]Page, error) {
	file, r, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	var pages []Page
	numPages := r.NumPage()
	for i := 1; i <= numPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from page %d: %w", i, err)
		}
		pages = append(pages, Page{
			Number:  i,
			Content: text,
		})
	}
	return pages, nil
}
