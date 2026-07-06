package menu

import (
	"fmt"

	"github.com/go-telegram/bot/models"

	"github.com/flykby/anonimus_chat/internal/bot/locales"
	"github.com/flykby/anonimus_chat/internal/shared"
)

const (
	CBBack            = "menu:back"
	CBStartChat       = "menu:start_chat"
	CBProfile         = "menu:profile"
	CBRules           = "menu:rules"
	CBEndDialog       = "menu:end_dialog"
	CBEndConfirm      = "end:confirm"
	CBEndCancel       = "end:cancel"
	CBQueueCancel     = "menu:queue_cancel"
	CBP2PReport       = "p2p:report"
	CBP2PBlock        = "p2p:block"
	CBProfilePremium  = "menu:profile:premium"
	CBProfileEdit     = "menu:profile:edit"
	CBProfileLanguage = "menu:profile:language"
	CBProfileDelete   = "menu:profile:delete"
	CBPremiumBuy      = "premium:buy"
)

type Action int

const (
	ActionUnknown Action = iota
	ActionStartChat
	ActionProfile
	ActionRules
	ActionEndDialog
	ActionCancelQueue
)

type Labels struct {
	MenuTitle             string
	StartChat             string
	Profile               string
	Rules                 string
	EndDialog             string
	Back                  string
	StartChatError        string
	StartChatActive       string
	QueueCancel           string
	QueueMatched          string
	QueueTimeout          string
	QueueCancelled        string
	ProfileBuyPremium     string
	ProfileExtendPremium  string
	ProfileEdit           string
	ProfileChangeLanguage string
	ProfileDelete         string
	ProfilePremiumStub    string
	ProfileEditStub       string
	ProfileLanguageStub   string
	ProfileDeleteStub     string
	EndDialogConfirm      string
	EndDialogConfirmYes   string
	EndDialogConfirmNo    string
	EndDialogCancelled    string
	EndDialogEnded        string
	PartnerEndedDialog    string
	PartnerBlockedDialog  string
	P2PDisallowedMsg      string
	P2PRateLimited        string
	P2PPhotoLimit         string
	P2PRelayError         string
	P2PReport             string
	P2PBlock              string
	P2PReportSent         string
	P2PBlocked            string
	P2PModerationHint     string
	DialogActiveHint      string
}

func LabelsFor(lang shared.Language) Labels {
	return Labels{
		MenuTitle:             locales.T("menu.title", lang, nil),
		StartChat:             locales.T("menu.start_chat", lang, nil),
		Profile:               locales.T("menu.profile", lang, nil),
		Rules:                 locales.T("menu.rules", lang, nil),
		EndDialog:             locales.T("menu.end_dialog", lang, nil),
		Back:                  locales.T("menu.back", lang, nil),
		StartChatError:        locales.T("menu.start_chat_error", lang, nil),
		StartChatActive:       locales.T("menu.start_chat_active", lang, nil),
		QueueCancel:           locales.T("menu.queue_cancel", lang, nil),
		QueueMatched:          locales.T("menu.queue_matched", lang, nil),
		QueueTimeout:          locales.T("menu.queue_timeout", lang, nil),
		QueueCancelled:        locales.T("menu.queue_cancelled", lang, nil),
		ProfileBuyPremium:     locales.T("profile.buy_premium", lang, nil),
		ProfileExtendPremium:  locales.T("profile.extend_premium", lang, nil),
		ProfileEdit:           locales.T("profile.edit", lang, nil),
		ProfileChangeLanguage: locales.T("profile.change_language", lang, nil),
		ProfileDelete:         locales.T("profile.delete", lang, nil),
		ProfilePremiumStub:    locales.T("profile.premium_stub", lang, nil),
		ProfileEditStub:       locales.T("profile.edit_stub", lang, nil),
		ProfileLanguageStub:   locales.T("profile.language_stub", lang, nil),
		ProfileDeleteStub:     locales.T("profile.delete_stub", lang, nil),
		EndDialogConfirm:      locales.T("dialog.end_confirm", lang, nil),
		EndDialogConfirmYes:   locales.T("dialog.end_confirm_yes", lang, nil),
		EndDialogConfirmNo:    locales.T("dialog.end_confirm_no", lang, nil),
		EndDialogCancelled:    locales.T("dialog.end_cancelled", lang, nil),
		EndDialogEnded:        locales.T("dialog.end_ended", lang, nil),
		PartnerEndedDialog:    locales.T("dialog.partner_ended", lang, nil),
		PartnerBlockedDialog:  locales.T("dialog.partner_blocked", lang, nil),
		P2PDisallowedMsg:      locales.T("p2p.disallowed", lang, nil),
		P2PRateLimited:        locales.T("p2p.rate_limited", lang, nil),
		P2PPhotoLimit:         locales.T("p2p.photo_limit", lang, nil),
		P2PRelayError:         locales.T("p2p.relay_error", lang, nil),
		P2PReport:             locales.T("p2p.report", lang, nil),
		P2PBlock:              locales.T("p2p.block", lang, nil),
		P2PReportSent:         locales.T("p2p.report_sent", lang, nil),
		P2PBlocked:            locales.T("p2p.blocked", lang, nil),
		P2PModerationHint:     locales.T("p2p.moderation_hint", lang, nil),
		DialogActiveHint:      locales.T("menu.dialog_active_hint", lang, nil),
	}
}

func ParseLanguage(lang string) shared.Language {
	switch shared.Language(lang) {
	case shared.LanguageEN:
		return shared.LanguageEN
	default:
		return shared.LanguageRU
	}
}

func ActionForText(text string) (Action, shared.Language) {
	for _, lang := range []shared.Language{shared.LanguageRU, shared.LanguageEN} {
		labels := LabelsFor(lang)
		switch text {
		case labels.StartChat:
			return ActionStartChat, lang
		case labels.Profile:
			return ActionProfile, lang
		case labels.Rules:
			return ActionRules, lang
		case labels.EndDialog:
			return ActionEndDialog, lang
		case labels.QueueCancel:
			return ActionCancelQueue, lang
		}
	}
	return ActionUnknown, shared.LanguageRU
}

func MainKeyboard(labels Labels) models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: labels.StartChat}},
			{{Text: labels.Profile}, {Text: labels.Rules}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}

func DialogKeyboard(labels Labels) models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: labels.EndDialog}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}

func ProfileSummary(age int16, gender, seeking string, lang shared.Language) string {
	return locales.T("profile.summary.template", lang, map[string]string{
		"age":     fmt.Sprint(age),
		"gender":  locales.ProfileGenderLabel(shared.Gender(gender), lang),
		"seeking": locales.SeekingLabel(shared.Gender(seeking), lang),
	})
}

func EndConfirmButtons(labels Labels) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: labels.EndDialogConfirmYes, CallbackData: CBEndConfirm},
			{Text: labels.EndDialogConfirmNo, CallbackData: CBEndCancel},
		},
	}
}

func P2PModerationButtons(labels Labels) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: labels.P2PReport, CallbackData: CBP2PReport},
			{Text: labels.P2PBlock, CallbackData: CBP2PBlock},
		},
	}
}

func BackButton(labels Labels) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: labels.Back, CallbackData: CBBack}
}

func RemoveKeyboard() models.ReplyKeyboardRemove {
	return models.ReplyKeyboardRemove{RemoveKeyboard: true}
}
