package dialog

import "testing"

func TestNormalizeRelayPayloadText(t *testing.T) {
	t.Parallel()

	content, text, fileID, err := normalizeRelayPayload(RelayRequest{
		Kind: RelayKindText,
		Text: " hello ",
	})
	if err != nil || content != "hello" || text != "hello" || fileID != "" {
		t.Fatalf("got content=%q text=%q file=%q err=%v", content, text, fileID, err)
	}
}

func TestNormalizeRelayPayloadPhoto(t *testing.T) {
	t.Parallel()

	content, text, fileID, err := normalizeRelayPayload(RelayRequest{
		Kind:           RelayKindPhoto,
		TelegramFileID: "file123",
	})
	if err != nil || content != "photo:file123" || text != "" || fileID != "file123" {
		t.Fatalf("got content=%q text=%q file=%q err=%v", content, text, fileID, err)
	}
}

func TestNormalizeRelayPayloadRejectsEmptyText(t *testing.T) {
	t.Parallel()

	_, _, _, err := normalizeRelayPayload(RelayRequest{
		Kind: RelayKindText,
		Text: "   ",
	})
	if err != ErrInvalidRelay {
		t.Fatalf("err = %v, want ErrInvalidRelay", err)
	}
}
