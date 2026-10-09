package infrastructure

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestResponseError(t *testing.T) {
	t.Run("ステータスと本文を含める", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(`{"error":"invalid_grant"}`))}

		err := responseError("token exchange", resp)

		want := `token exchange failed (status 400): {"error":"invalid_grant"}`
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	})

	t.Run("長い本文は先頭だけを含める", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", maxErrorBodyBytes*2)))}

		err := responseError("userinfo request", resp)

		prefix := "userinfo request failed (status 502): "
		if got := len(err.Error()) - len(prefix); got != maxErrorBodyBytes {
			t.Errorf("body length in error = %d, want %d", got, maxErrorBodyBytes)
		}
	})

	t.Run("本文を読めなくても、ステータスと読み取りの失敗を返す", func(t *testing.T) {
		readErr := errors.New("connection reset")
		resp := &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(failingReader{err: readErr})}

		err := responseError("emails request", resp)

		if !errors.Is(err, readErr) {
			t.Errorf("error = %v, want to wrap %v", err, readErr)
		}
		if !strings.Contains(err.Error(), "status 500") {
			t.Errorf("error = %q, want to contain the status", err.Error())
		}
	})
}
