package rules

import (
	"embed"
	"fmt"
	"strings"

	"github.com/flykby/anonimus_chat/internal/shared"
)

//go:embed rules_ru.md rules_en.md
var files embed.FS

const RulesVersion = "v1"
const maxChunkLen = 4096

func Messages(lang shared.Language) ([]string, error) {
	name := "rules_ru.md"
	if lang == shared.LanguageEN {
		name = "rules_en.md"
	}

	raw, err := files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read rules %s: %w", name, err)
	}

	text := strings.TrimSpace(string(raw))
	if text == "" {
		return nil, fmt.Errorf("rules %s is empty", name)
	}
	return splitChunks(text, maxChunkLen), nil
}

func splitChunks(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(current.String()))
		current.Reset()
	}

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		sep := "\n\n"
		candidate := paragraph
		if current.Len() > 0 {
			candidate = current.String() + sep + paragraph
		}

		if len(candidate) <= maxLen {
			current.Reset()
			current.WriteString(strings.TrimSpace(candidate))
			continue
		}

		flush()
		if len(paragraph) <= maxLen {
			current.WriteString(paragraph)
			continue
		}

		for len(paragraph) > maxLen {
			chunks = append(chunks, paragraph[:maxLen])
			paragraph = paragraph[maxLen:]
		}
		if strings.TrimSpace(paragraph) != "" {
			current.WriteString(strings.TrimSpace(paragraph))
		}
	}

	flush()
	return chunks
}
