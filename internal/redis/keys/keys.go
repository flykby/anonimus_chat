package keys

import (
	"fmt"

	"github.com/flykby/anonimus_chat/internal/shared"
)

const Prefix = "anonimus"

func P2PQueue(gender shared.Gender) string {
	return fmt.Sprintf("%s:queue:p2p:%s", Prefix, gender)
}

func Session(userID int64) string {
	return fmt.Sprintf("%s:session:%d", Prefix, userID)
}

func FSM(telegramID int64) string {
	return fmt.Sprintf("%s:fsm:%d", Prefix, telegramID)
}

func RateLimit(userID int64, action string) string {
	return fmt.Sprintf("%s:ratelimit:%d:%s", Prefix, userID, action)
}

func DialogContext(dialogID int64) string {
	return fmt.Sprintf("%s:dialog_ctx:%d", Prefix, dialogID)
}
