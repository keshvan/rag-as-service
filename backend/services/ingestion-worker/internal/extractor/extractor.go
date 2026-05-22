package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
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
		return extractDOCX(localPath)
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

func extractDOCX(path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}
	defer reader.Close()

	var documentXML *zip.File
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			documentXML = file
			break
		}
	}
	if documentXML == nil {
		return "", fmt.Errorf("docx word/document.xml not found")
	}

	rc, err := documentXML.Open()
	if err != nil {
		return "", fmt.Errorf("open docx document.xml: %w", err)
	}
	defer rc.Close()

	text, err := extractDOCXDocumentText(rc)
	if err != nil {
		return "", err
	}

	text = normalizeText(text)
	if text == "" {
		return "", fmt.Errorf("docx contains no extractable text")
	}

	return text, nil
}

func extractDOCXDocumentText(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)

	var out strings.Builder
	var paragraph strings.Builder
	inText := false

	flushParagraph := func() {
		text := strings.TrimSpace(paragraph.String())
		if text != "" {
			if out.Len() > 0 {
				out.WriteByte('\n')
			}
			out.WriteString(text)
		}
		paragraph.Reset()
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("parse docx document.xml: %w", err)
		}

		switch tok := token.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "t":
				inText = true
			case "tab":
				paragraph.WriteByte('\t')
			case "br", "cr":
				paragraph.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				paragraph.Write(tok)
			}
		case xml.EndElement:
			switch tok.Name.Local {
			case "t":
				inText = false
			case "p":
				flushParagraph()
			}
		}
	}

	flushParagraph()

	return out.String(), nil
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
