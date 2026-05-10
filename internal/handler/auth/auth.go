package auth

import (
	"errors"
	"net/http"
	"strconv"
)

// HeaderUserID は本人を表すユーザーIDを運ぶリクエストヘッダーのキー。
// 学習用に「クライアントが userId をヘッダーで自己申告する」スタイルを採用している。
// 後で JWT 等の認証層を入れる時は、この関数の中身だけ差し替えれば呼び出し側はそのまま使える。
const HeaderUserID = "X-User-Id"

// ErrMissingUserID は X-User-Id ヘッダーが付いていない、または値が空の場合に返る。
var ErrMissingUserID = errors.New("missing X-User-Id header")

// ErrInvalidUserID は X-User-Id ヘッダーの値が正の整数として解釈できない場合に返る。
var ErrInvalidUserID = errors.New("invalid X-User-Id header")

// UserIDFromHeader は X-User-Id ヘッダーから本人のユーザーIDを取り出す。
// 認証層を入れる時はこの関数を「トークンを検証して userId を取り出す」実装に置き換える。
func UserIDFromHeader(r *http.Request) (int64, error) {
	raw := r.Header.Get(HeaderUserID)
	if raw == "" {
		return 0, ErrMissingUserID
	}
	userID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidUserID
	}
	return userID, nil
}
