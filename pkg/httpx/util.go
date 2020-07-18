package httpx

import "net/http"

func CloseResponse(r *http.Response, _ ...interface{}) {
	if r != nil && r.Body != nil {
		_ = r.Body.Close()
	}
}
