package localization

import (
	"strings"
	"sync"
)

var DefaultTranslator *Translator

func init() {
	DefaultTranslator = NewTranslator()
}

type Translator struct {
	sync.RWMutex
	languages map[string]*Language
}

func NewTranslator() *Translator {
	return &Translator{
		languages: make(map[string]*Language),
	}
}

func (t *Translator) Languages() []string {
	t.Lock()

	ln := len(t.languages)
	values := make([]string, ln)

	i := 0
	for k := range t.languages {
		values[i] = k
		i++
	}

	t.Unlock()

	return values
}

func (t *Translator) Add(language, key, value string) {
	t.Lock()
	lan, ok := t.languages[language]
	if !ok {
		lan = NewLanguage()
		t.languages[language] = lan
	}
	t.Unlock()

	lan.AddTranslation(key, value)
}

func (t *Translator) Translate(language, key string) (string, bool) {
	if language == "" {
		return FormatKey(key), false
	}

	t.RLock()
	lan, ok := t.languages[language]
	t.RUnlock()

	if !ok {
		return FormatKey(key), false
	}

	return lan.Translate(key)
}

type Language struct {
	sync.RWMutex
	Translations map[string]string
}

func NewLanguage() *Language {
	return &Language{
		Translations: make(map[string]string),
	}
}

func (lan *Language) AddTranslation(key, value string) {
	lan.Lock()
	lan.Translations[key] = value
	lan.Unlock()

}

func (lan *Language) Translate(key string) (string, bool) {
	lan.RLock()
	defer lan.RUnlock()

	translation, ok := lan.Translations[key]
	if !ok {
		translation = FormatKey(translation)
	}
	return translation, ok
}

func FormatKey(v string) string {
	v = strings.TrimPrefix(v, "@@")
	return v
}
