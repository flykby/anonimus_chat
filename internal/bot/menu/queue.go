package menu

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/shared"
)

func QueueWaitingText(count int64, gender shared.Gender, lang shared.Language) string {
	seekingKey := "queue.seeking_male"
	if gender == shared.GenderFemale && lang == shared.LanguageRU {
		seekingKey = "queue.seeking_female"
	}
	return locales.T("queue.waiting", lang, map[string]string{
		"count":   fmt.Sprint(count),
		"seeking": locales.T(seekingKey, lang, nil),
	})
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
