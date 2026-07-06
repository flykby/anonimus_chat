package profile

import (
	"context"
	"errors"

	"github.com/flykby/anonimus_chat/internal/db"
	"github.com/flykby/anonimus_chat/internal/dialog"
	"github.com/flykby/anonimus_chat/internal/match"
	"github.com/flykby/anonimus_chat/internal/redis/session"
)

type DeleteService struct {
	Users     *db.UsersRepo
	Dialogs   *db.DialogsRepo
	DialogSvc *dialog.Service
	MatchSvc  *match.Service
	Sessions  *session.Store
	Benefits  *db.DeletionBenefitsRepo
}

type DeleteResult struct {
	PartnerTelegramID *int64
	PartnerLanguage   *string
}

func (s *DeleteService) Delete(ctx context.Context, telegramID int64) (DeleteResult, error) {
	up, ok, err := s.Users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return DeleteResult{}, err
	}
	if !ok {
		return DeleteResult{}, db.ErrUserNotFound
	}

	var result DeleteResult

	if s.MatchSvc != nil {
		_ = s.MatchSvc.Cancel(ctx, telegramID)
	}

	if active, err := s.Users.HasActiveDialog(ctx, up.User.ID); err != nil {
		return DeleteResult{}, err
	} else if active && s.Dialogs != nil && s.DialogSvc != nil {
		d, found, err := s.Dialogs.GetActiveByUserID(ctx, up.User.ID)
		if err != nil {
			return DeleteResult{}, err
		}
		if found {
			endResp, err := s.DialogSvc.End(ctx, dialog.EndRequest{
				DialogID: d.ID,
				UserID:   up.User.ID,
				Reason:   "profile_deleted",
			})
			if err != nil && !errors.Is(err, dialog.ErrDialogNotFound) {
				return DeleteResult{}, err
			}
			result.PartnerTelegramID = endResp.PartnerTelegramID
			result.PartnerLanguage = endResp.PartnerLanguage
		}
	}

	if s.Sessions != nil {
		_ = s.Sessions.Delete(ctx, up.User.ID)
	}

	if err := s.Users.SoftDelete(ctx, telegramID, "user_requested"); err != nil {
		return DeleteResult{}, err
	}

	return result, nil
}

// EligibleForFreeUnlock is reserved for the adult-photo anti-abuse offer (task 021).
func (s *DeleteService) EligibleForFreeUnlock(ctx context.Context, telegramID int64) (bool, error) {
	if s.Benefits == nil {
		return false, nil
	}
	used, err := s.Benefits.FreeUnlockUsed(ctx, telegramID)
	if err != nil || used {
		return false, err
	}
	return false, nil
}
