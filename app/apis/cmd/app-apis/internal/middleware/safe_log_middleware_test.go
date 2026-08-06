package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSafeLogPreservesRequestBody(t *testing.T) {
	const body = `{"phone":"13800138000","password":"password123"}`
	handler := SafeLog(func(w http.ResponseWriter, r *http.Request) {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if string(content) != body {
			t.Fatalf("request body = %q, want %q", content, body)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	handler(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
