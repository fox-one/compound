package render

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Response internal error msg as hint
var ResponseErrorMessageAsHint bool

func init() {
	v := os.Getenv("RESPONSE_ERROR_MESSAGE_AS_HINT")
	ResponseErrorMessageAsHint, _ = strconv.ParseBool(v)
}

type wrapResponse struct {
	status int
	header http.Header
	buf    *bytes.Buffer
}

func (w *wrapResponse) Header() http.Header {
	return w.header
}

func (w *wrapResponse) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *wrapResponse) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func (w *wrapResponse) isJsonContent() bool {
	typ := w.header.Get("Content-Type")
	return strings.HasPrefix(typ, "application/json")
}

type dataResponse struct {
	Data json.RawMessage `json:"data,omitempty"`
}

type errorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Hint string `json:"hint,omitempty"`
}
