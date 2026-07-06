package locales

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/flykby/anonimus_chat/internal/shared"
)

//go:embed ru.yaml en.yaml
var files embed.FS

var (
	once    sync.Once
	catalog map[shared.Language]map[string]string
)

func initCatalog() {
	catalog = map[shared.Language]map[string]string{
		shared.LanguageRU: mustLoad("ru.yaml"),
		shared.LanguageEN: mustLoad("en.yaml"),
	}
}

func mustLoad(name string) map[string]string {
	raw, err := files.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("locales: read %s: %v", name, err))
	}
	var nested map[string]any
	if err := yaml.Unmarshal(raw, &nested); err != nil {
		panic(fmt.Sprintf("locales: parse %s: %v", name, err))
	}
	out := make(map[string]string)
	flatten("", nested, out)
	return out
}

func flatten(prefix string, node map[string]any, out map[string]string) {
	for key, value := range node {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]any:
			flatten(fullKey, v, out)
		case string:
			out[fullKey] = v
		default:
			panic(fmt.Sprintf("locales: unsupported value type for key %q", fullKey))
		}
	}
}

// T returns a localized string for key and language. Missing EN keys fall back to RU.
func T(key string, lang shared.Language, params map[string]string) string {
	once.Do(initCatalog)

	text, ok := lookup(key, lang)
	if !ok {
		return key
	}
	if params == nil {
		return text
	}
	for name, value := range params {
		text = strings.ReplaceAll(text, "{"+name+"}", value)
	}
	return text
}

func lookup(key string, lang shared.Language) (string, bool) {
	if lang == shared.LanguageEN {
		if text, ok := catalog[shared.LanguageEN][key]; ok {
			return text, true
		}
	}
	if text, ok := catalog[shared.LanguageRU][key]; ok {
		return text, true
	}
	return "", false
}

func GenderLabel(g shared.Gender, lang shared.Language) string {
	switch g {
	case shared.GenderMale:
		return T("gender.male", lang, nil)
	case shared.GenderFemale:
		return T("gender.female", lang, nil)
	default:
		return string(g)
	}
}

func SeekingLabel(g shared.Gender, lang shared.Language) string {
	switch g {
	case shared.GenderMale:
		return T("gender.seeking_male", lang, nil)
	case shared.GenderFemale:
		return T("gender.seeking_female", lang, nil)
	default:
		return string(g)
	}
}

func ProfileGenderLabel(g shared.Gender, lang shared.Language) string {
	if lang == shared.LanguageEN {
		switch g {
		case shared.GenderMale:
			return T("gender.guy", lang, nil)
		case shared.GenderFemale:
			return T("gender.girl", lang, nil)
		}
	}
	return GenderLabel(g, lang)
}

func LanguageName(l shared.Language) string {
	switch l {
	case shared.LanguageEN:
		return T("language.en", shared.LanguageEN, nil)
	default:
		return T("language.ru", shared.LanguageRU, nil)
	}
}

func LanguageCode(l shared.Language) string {
	switch l {
	case shared.LanguageEN:
		return T("language.code_en", shared.LanguageEN, nil)
	default:
		return T("language.code_ru", shared.LanguageRU, nil)
	}
}
