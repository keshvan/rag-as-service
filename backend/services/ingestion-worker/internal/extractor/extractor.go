package extractor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(ctx context.Context, localPath string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	switch strings.ToLower(filepath.Ext(localPath)) {
	case ".pdf":
		return extractPDF(localPath)
	case ".txt":
		return extractTXT(localPath)
	case ".docx":
		return "", fmt.Errorf("docx extraction is not implemented yet")
	default:
		return "", fmt.Errorf("unsupported file type: %s", filepath.Ext(localPath))
	}
}

func extractPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract plain text from pdf: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return "", fmt.Errorf("read pdf text: %w", err)
	}

	text := normalizeText(buf.String())
	if text == "" {
		return "", fmt.Errorf("pdf contains no extractable text")
	}

	return text, nil
}

func extractTXT(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read txt: %w", err)
	}

	text := normalizeText(string(data))
	if text == "" {
		return "", fmt.Errorf("txt is empty")
	}

	return text, nil
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}

	return strings.TrimSpace(strings.Join(out, "\n"))
}
