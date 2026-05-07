package chunking

import (
	"strings"

	"github.com/keshvan/rag-as-service/backend/pkg/common/qdrant"
)

type Chunker struct {
	size    int
	overlap int
}

func NewChunker(size, overlap int) *Chunker {
	if size <= 0 {
		size = 1000
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 5
	}

	return &Chunker{
		size:    size,
		overlap: overlap,
	}
}

func (c *Chunker) Split(text string) []qdrant.Chunk {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	runes := []rune(text)
	if len(runes) <= c.size {
		return []qdrant.Chunk{{Index: 0, Text: string(runes)}}
	}

	step := c.size - c.overlap
	if step <= 0 {
		step = c.size
	}

	minTailSize := c.size / 5
	if minTailSize < 100 {
		minTailSize = 100
	}

	var out []qdrant.Chunk
	start := 0
	idx := 0

	for start < len(runes) {
		end := start + c.size
		if end > len(runes) {
			end = len(runes)
		}

		// Если не конец текста — ищем ближайший пробел/перенос назад, чтобы не резать слово
		if end < len(runes) {
			searchWindow := 100
			if searchWindow > c.size {
				searchWindow = c.size / 4
			}
			searchStart := end - searchWindow
			if searchStart < start {
				searchStart = start
			}
			for i := end - 1; i > searchStart; i-- {
				if runes[i] == ' ' || runes[i] == '\n' || runes[i] == '\t' {
					end = i
					break
				}
			}
		}

		chunkText := strings.TrimSpace(string(runes[start:end]))
		if chunkText != "" {
			out = append(out, qdrant.Chunk{Index: idx, Text: chunkText})
			idx++
		}

		if end >= len(runes) {
			break
		}

		nextStart := end - c.overlap
		if nextStart <= start {
			nextStart = start + step
			if nextStart >= end {
				nextStart = end
			}
		}
		start = nextStart
	}

	// Слишком короткий последний чанк — вливаем в предыдущий
	if len(out) >= 2 {
		last := out[len(out)-1]
		if len([]rune(last.Text)) < minTailSize {
			prev := out[len(out)-2]
			merged := strings.TrimSpace(prev.Text + " " + last.Text)
			out = out[:len(out)-1]
			out[len(out)-1].Text = merged
		}
	}

	for i := range out {
		out[i].Index = i
	}

	return out
}
