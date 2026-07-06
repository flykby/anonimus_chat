package menu

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/shared"
)

func QueueWaitingText(count int64, gender shared.Gender, lang shared.Language) string {
	if lang == shared.LanguageEN {
		return fmt.Sprintf(
			"You are in the queue looking for a partner.\n\nUsers matching your preferences: %d\n\n%s",
			count,
			queueSeekingEN(),
		)
	}
	return fmt.Sprintf(
		"Вы находитесь в очереди на поиск собеседника.\n\nПользователей подходящих под ваши параметры: %d\n\n%s",
		count,
		queueSeekingRU(gender),
	)
}

func queueSeekingRU(gender shared.Gender) string {
	if gender == shared.GenderFemale {
		return "Ищем лучшую из них"
	}
	return "Ищем лучшего из них"
}

func queueSeekingEN() string {
	return "Looking for the best match"
}

func QueueWaitingKeyboard(labels Labels) models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: labels.QueueCancel}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}
