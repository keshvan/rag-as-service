package extractor

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDOCX(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "sample.docx")
	writeMinimalDOCX(t, path, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r><w:t>Первый абзац</w:t></w:r>
    </w:p>
    <w:p>
      <w:r><w:t>Второй</w:t></w:r>
      <w:r><w:t> абзац</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`)

	text, err := NewExtractor().Extract(context.Background(), path)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	want := "Первый абзац\nВторой абзац"
	if text != want {
		t.Fatalf("unexpected text:\nwant: %q\n got: %q", want, text)
	}
}

func TestExtractDOCXRejectsEmptyDocument(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "empty.docx")
	writeMinimalDOCX(t, path, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body><w:p><w:r><w:t>   </w:t></w:r></w:p></w:body>
</w:document>`)

	_, err := NewExtractor().Extract(context.Background(), path)
	if err == nil {
		t.Fatal("expected empty docx error")
	}
	if !strings.Contains(err.Error(), "docx contains no extractable text") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeMinimalDOCX(t *testing.T, path, documentXML string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	defer archive.Close()

	contentTypes, err := archive.Create("[Content_Types].xml")
	if err != nil {
		t.Fatalf("create content types: %v", err)
	}
	if _, err := contentTypes.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`)); err != nil {
		t.Fatalf("write content types: %v", err)
	}

	doc, err := archive.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create document.xml: %v", err)
	}
	if _, err := doc.Write([]byte(documentXML)); err != nil {
		t.Fatalf("write document.xml: %v", err)
	}
}
