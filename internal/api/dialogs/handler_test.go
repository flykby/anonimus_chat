package dialogapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEndInvalidDialogID(t *testing.T) {
	t.Parallel()

	h := &Handler{Dialogs: nil, Users: nil}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(endRequest{TelegramID: 1, Reason: "user_confirmed"})
	req := httptest.NewRequest(http.MethodPost, "/dialogs/abc/end", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
