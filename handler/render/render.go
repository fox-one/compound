package render

import (
	"encoding/json"
	"net/http"
)

type H map[string]interface{}

// JSON render with json
func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

// Text render with text
func Text(w http.ResponseWriter, t string) {
	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(t))
}

// Error write error
func Error(w http.ResponseWriter, statusCode, errCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	enc := json.NewEncoder(w)
	_ = enc.Encode(H{"code": errCode, "msg": err.Error()})
}

// BadRequest bad request error
func BadRequest(w http.ResponseWriter, err error) {
	Error(w, http.StatusBadRequest, -1, err)
}

// NotFoundRequest not found request error
func NotFoundRequest(w http.ResponseWriter, err error) {
	Error(w, http.StatusNotFound, -1, err)
}
