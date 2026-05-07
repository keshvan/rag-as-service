package chunking

import "strings"

type Chunk struct {
	Index int
	Text  string
}

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

func (c *Chunker) Split(text string) []Chunk {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	runes := []rune(text)
	step := c.size - c.overlap
	if step <= 0 {
		step = c.size
	}

	out := make([]Chunk, 0, max(1, len(runes)/step))
	for start, idx := 0, 0; start < len(runes); start, idx = start+step, idx+1 {
		end := start + c.size
		if end > len(runes) {
			end = len(runes)
		}

		chunkText := strings.TrimSpace(string(runes[start:end]))
		if chunkText != "" {
			out = append(out, Chunk{
				Index: idx,
				Text:  chunkText,
			})
		}

		if end == len(runes) {
			break
		}
	}

	return out
}
