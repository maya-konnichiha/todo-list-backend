package errhandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Response はエラーレスポンスの JSON ボディ。`{"code": "...", "message": "..."}` 形式。
type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Write は指定ステータスでエラー JSON を書き出す。
func Write(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(Response{Code: code, Message: message}); err != nil {
		slog.Error("failed to encode error response", slog.Any("error", err))
	}
}

// BadRequest は 400 を返す。
func BadRequest(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusBadRequest, code, message)
}

// NotFound は 404 を返す。
func NotFound(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusNotFound, code, message)
}

// Conflict は 409 を返す。
func Conflict(w http.ResponseWriter, code, message string) {
	Write(w, http.StatusConflict, code, message)
}

// Internal は 500 を返し、元のエラーは slog に記録する(クライアントには露出させない)。
func Internal(w http.ResponseWriter, err error) {
	slog.Error("internal server error", slog.Any("error", err))
	Write(w, http.StatusInternalServerError, "INTERNAL_ERROR", "内部エラーが発生しました")
}
