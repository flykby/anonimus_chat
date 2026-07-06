package deleteprofile

import (
	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	CBCancel   = "delete:cancel"
	CBConfirm1 = "delete:yes1"
	CBConfirm2 = "delete:yes2"
)

func Confirm1(lang shared.Language) string {
	return locales.T("profile.delete_flow.confirm1", lang, nil)
}

func Confirm2(lang shared.Language) string {
	return locales.T("profile.delete_flow.confirm2", lang, nil)
}

func Cancelled(lang shared.Language) string {
	return locales.T("profile.delete_flow.cancelled", lang, nil)
}

func Done(lang shared.Language) string {
	return locales.T("profile.delete_flow.done", lang, nil)
}

func Error(lang shared.Language) string {
	return locales.T("profile.delete_flow.error", lang, nil)
}

func ButtonConfirm(lang shared.Language) string {
	return locales.T("profile.delete_flow.button_confirm", lang, nil)
}

func ButtonFinalConfirm(lang shared.Language) string {
	return locales.T("profile.delete_flow.button_final_confirm", lang, nil)
}

func ButtonCancel(lang shared.Language) string {
	return locales.T("profile.delete_flow.button_cancel", lang, nil)
}

func Confirm1Buttons(lang shared.Language) [][]models.InlineKeyboardButton {
	return confirmButtons(lang, CBConfirm1)
}

func Confirm2Buttons(lang shared.Language) [][]models.InlineKeyboardButton {
	return confirmButtons(lang, CBConfirm2)
}

func confirmButtons(lang shared.Language, confirmData string) [][]models.InlineKeyboardButton {
	confirmLabel := ButtonConfirm(lang)
	if confirmData == CBConfirm2 {
		confirmLabel = ButtonFinalConfirm(lang)
	}
	return [][]models.InlineKeyboardButton{
		{
			{Text: confirmLabel, CallbackData: confirmData},
			{Text: ButtonCancel(lang), CallbackData: CBCancel},
		},
	}
}
