package registration

import (
	"fmt"
	"strconv"
	"strings"

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

const (
	WelcomeText    = "Привет! Добро пожаловать в анонимный чат.\n\nЧтобы начать, заполни короткую анкету."
	AgePrompt      = "Сколько тебе лет?\n\nНапиши число от 18 до 99."
	AgeTooYoung    = "К сожалению, сервис доступен только с 18 лет."
	AgeInvalid     = "Введи возраст числом от 18 до 99."
	GenderPrompt   = "Твой пол:"
	SeekingPrompt  = "Кого ты ищешь?"
	LanguagePrompt = "Выбери язык интерфейса:"
	UseButtonsHint = "Пожалуйста, выбери вариант кнопкой под сообщением."
	MainMenuStub   = "Регистрация завершена! Главное меню скоро будет здесь."
)

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

func GenderLabel(g shared.Gender) string {
	switch g {
	case shared.GenderMale:
		return "Парень"
	case shared.GenderFemale:
		return "Девушка"
	default:
		return string(g)
	}
}

func SeekingLabel(g shared.Gender) string {
	switch g {
	case shared.GenderMale:
		return "Парня"
	case shared.GenderFemale:
		return "Девушку"
	default:
		return string(g)
	}
}

func LanguageLabel(l shared.Language) string {
	switch l {
	case shared.LanguageRU:
		return "Русский"
	case shared.LanguageEN:
		return "English"
	default:
		return string(l)
	}
}

func ConfirmationText(d regdraft.Draft) string {
	return fmt.Sprintf(
		"Проверь анкету:\n\nВозраст: %d\nПол: %s\nИщу: %s\nЯзык: %s\n\nВсё верно?",
		d.Age,
		GenderLabel(d.Gender),
		SeekingLabel(d.Seeking),
		LanguageLabel(d.Language),
	)
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
