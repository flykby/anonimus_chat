package events

import "fmt"

type Type string

const (
	TypeUserRegistered     Type = "user.registered"
	TypeUserProfileUpdated Type = "user.profile_updated"
	TypeUserDeleted        Type = "user.deleted"

	TypeDialogStarted Type = "dialog.started"
	TypeDialogEnded   Type = "dialog.ended"

	TypeMessageSent     Type = "message.sent"
	TypeMessageReceived Type = "message.received"

	TypePhotoRequested Type = "photo.requested"
	TypePhotoSent      Type = "photo.sent"
	TypePhotoUnlocked  Type = "photo.unlocked"

	TypePremiumPurchased Type = "premium.purchased"
	TypePremiumExpired   Type = "premium.expired"

	TypeQueueEntered Type = "queue.entered"
	TypeQueueMatched Type = "queue.matched"
)

var allTypes = []Type{
	TypeUserRegistered,
	TypeUserProfileUpdated,
	TypeUserDeleted,
	TypeDialogStarted,
	TypeDialogEnded,
	TypeMessageSent,
	TypeMessageReceived,
	TypePhotoRequested,
	TypePhotoSent,
	TypePhotoUnlocked,
	TypePremiumPurchased,
	TypePremiumExpired,
	TypeQueueEntered,
	TypeQueueMatched,
}

func (t Type) Valid() bool {
	for _, known := range allTypes {
		if t == known {
			return true
		}
	}
	return false
}

func (t Type) String() string {
	return string(t)
}

func validateType(t Type) error {
	if !t.Valid() {
		return fmt.Errorf("unknown event type %q", t)
	}
	return nil
}
