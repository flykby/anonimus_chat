package registration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/redis/regdraft"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	StateAge      = "reg:age"
	StateGender   = "reg:gender"
	StateSeeking  = "reg:seeking"
	StateLanguage = "reg:language"
	StateConfirm  = "reg:confirm"

	CBStart          = "reg:start"
	CBGenderMale     = "reg:gender:male"
	CBGenderFemale   = "reg:gender:female"
	CBSeekingMale    = "reg:seeking:male"
	CBSeekingFemale  = "reg:seeking:female"
	CBLanguageRU     = "reg:lang:ru"
	CBLanguageEN     = "reg:lang:en"
	CBConfirmYes     = "reg:confirm:yes"
	CBConfirmRestart = "reg:confirm:restart"
)

func WelcomeText(lang shared.Language) string {
	return locales.T("registration.welcome", lang, nil)
}

func WelcomeHint(lang shared.Language) string {
	return locales.T("registration.welcome_hint", lang, nil)
}

func AgePrompt(lang shared.Language) string {
	return locales.T("registration.age.prompt", lang, nil)
}

func AgeTooYoung(lang shared.Language) string {
	return locales.T("registration.age.too_young", lang, nil)
}

func AgeInvalid(lang shared.Language) string {
	return locales.T("registration.age.invalid", lang, nil)
}

func GenderPrompt(lang shared.Language) string {
	return locales.T("registration.gender.prompt", lang, nil)
}

func SeekingPrompt(lang shared.Language) string {
	return locales.T("registration.seeking.prompt", lang, nil)
}

func LanguagePrompt(lang shared.Language) string {
	return locales.T("registration.language.prompt", lang, nil)
}

func UseButtonsHint(lang shared.Language) string {
	return locales.T("registration.use_buttons_hint", lang, nil)
}

func StartFormButton(lang shared.Language) string {
	return locales.T("registration.button.start_form", lang, nil)
}

func ConfirmYesButton(lang shared.Language) string {
	return locales.T("registration.button.confirm_yes", lang, nil)
}

func ConfirmRestartButton(lang shared.Language) string {
	return locales.T("registration.button.confirm_restart", lang, nil)
}

func GenderButtonMale(lang shared.Language) string {
	return locales.T("registration.button.gender_male", lang, nil)
}

func GenderButtonFemale(lang shared.Language) string {
	return locales.T("registration.button.gender_female", lang, nil)
}

func SeekingButtonMale(lang shared.Language) string {
	return locales.T("registration.button.seeking_male", lang, nil)
}

func SeekingButtonFemale(lang shared.Language) string {
	return locales.T("registration.button.seeking_female", lang, nil)
}

func LanguageButtonRU(lang shared.Language) string {
	return locales.T("registration.button.lang_ru", lang, nil)
}

func LanguageButtonEN(lang shared.Language) string {
	return locales.T("registration.button.lang_en", lang, nil)
}

func LoadDraftError(lang shared.Language) string {
	return locales.T("registration.error.load_draft", lang, nil)
}

func LoadDraftErrorShort(lang shared.Language) string {
	return locales.T("registration.error.load_draft_short", lang, nil)
}

func SaveProfileError(lang shared.Language) string {
	return locales.T("registration.error.save_profile", lang, nil)
}

func ParseAge(text string) (int16, error) {
	text = strings.TrimSpace(text)
	age, err := strconv.ParseInt(text, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("not a number")
	}
	if age < 18 || age > 99 {
		return 0, fmt.Errorf("out of range")
	}
	return int16(age), nil
}

func IsTooYoung(text string) bool {
	text = strings.TrimSpace(text)
	age, err := strconv.ParseInt(text, 10, 16)
	return err == nil && age > 0 && age < 18
}

func ConfirmationText(d regdraft.Draft) string {
	lang := d.Language
	if lang == "" {
		lang = shared.LanguageRU
	}
	return locales.T("registration.confirm.template", lang, map[string]string{
		"age":      fmt.Sprint(d.Age),
		"gender":   locales.GenderLabel(d.Gender, lang),
		"seeking":  locales.SeekingLabel(d.Seeking, lang),
		"language": locales.LanguageName(d.Language),
	})
}

func ParseGenderCallback(data string) (shared.Gender, bool) {
	switch data {
	case CBGenderMale, CBSeekingMale:
		return shared.GenderMale, true
	case CBGenderFemale, CBSeekingFemale:
		return shared.GenderFemale, true
	default:
		return "", false
	}
}

func ParseLanguageCallback(data string) (shared.Language, bool) {
	switch data {
	case CBLanguageRU:
		return shared.LanguageRU, true
	case CBLanguageEN:
		return shared.LanguageEN, true
	default:
		return "", false
	}
}

// RegLanguage returns draft language during registration or RU by default.
func RegLanguage(d regdraft.Draft, ok bool) shared.Language {
	if ok && d.Language != "" {
		return d.Language
	}
	return shared.LanguageRU
}
