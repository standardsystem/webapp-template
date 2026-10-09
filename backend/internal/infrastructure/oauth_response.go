package infrastructure

import (
	"fmt"
	"io"
	"net/http"
)

// maxErrorBodyBytes は、エラーメッセージに含めるレスポンス本文の上限です。
const maxErrorBodyBytes = 1024

// responseError は、プロバイダが 200 以外を返したときのエラーを作ります。
// 原因の手がかりとして本文の先頭を含めます。what は "token exchange" のような操作名です。
func responseError(what string, resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
	if err != nil {
		return fmt.Errorf("%s failed (status %d): failed to read response body: %w", what, resp.StatusCode, err)
	}
	return fmt.Errorf("%s failed (status %d): %s", what, resp.StatusCode, body)
}
