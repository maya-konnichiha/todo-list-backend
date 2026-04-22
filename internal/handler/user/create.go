package user

import (
	"encoding/json"
	"errors"
	"net/http"

	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// CreateHandler は POST /users を処理するハンドラ。
//
// 構造体にして usecase を保持するのは DI の流儀（Step 5 で述べた通り）。
// HTTP フレームワーク（Gin）を外した後も、この構造は変わらない。
// これは「フレームワークは handler 層のごく表層にしか影響しない」ことの証。
type CreateHandler struct {
	usecase *userusecase.CreateUser
}

// NewCreateHandler はコンストラクタ。
func NewCreateHandler(uc *userusecase.CreateUser) *CreateHandler {
	return &CreateHandler{usecase: uc}
}

// Handle は http.HandlerFunc のシグネチャ `func(w http.ResponseWriter, r *http.Request)` に合う。
//
// Gin 時代との対比:
//
//	Gin: func(c *gin.Context)          — c から全部取る
//	std: func(w http.ResponseWriter, r *http.Request) — リクエストとレスポンスが分離
//
// 標準では:
//   - 入力 → r（*http.Request）: Body, URL, Header, Context 等
//   - 出力 → w（http.ResponseWriter）: Header() + WriteHeader(status) + Write(body)
//
// Gin の便利メソッドは、結局これらのラッパーだったことがわかる。
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// ---- 1. JSON デコード ---------------------------------------------
	//
	// json.NewDecoder(r.Body).Decode(&req):
	//   - r.Body は io.ReadCloser。Decoder はストリームで読める（省メモリ）
	//   - フィールドの型が合わない / 壊れた JSON はここでエラー
	//   - 空 Body や無効な UTF-8 も弾かれる
	//
	// ※ ここでの err は「構文エラー / 型ミスマッチ」であり、
	//   「UserName が空」のような業務バリデーションではない。
	//   業務バリデーションは下の usecase 呼び出しで domain.NewUser が行う。
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// ---- 2. usecase 呼び出し ------------------------------------------
	//
	// r.Context() で Gin 時代の c.Request.Context() 相当を取得。
	// 実はこれが本来の形で、Gin の c.Request.Context() も同じものを返していた。
	created, err := h.usecase.Execute(r.Context(), userusecase.CreateUserParams{
		UserName:  req.UserName,
		UserEmail: req.UserEmail,
	})
	if err != nil {
		// ---- 3. エラーを HTTP ステータスに変換 ---------------------
		//
		// Gin 時代と全く同じ分岐ロジック。
		// sentinel エラーを使った設計は「フレームワーク非依存」で、
		// 今回の書き換えでも手を入れる必要がない。
		switch {
		case errors.Is(err, userdomain.ErrUserNameEmpty),
			errors.Is(err, userdomain.ErrUserNameTooLong),
			errors.Is(err, userdomain.ErrUserEmailEmpty),
			errors.Is(err, userdomain.ErrUserEmailInvalid):
			writeErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, userdomain.ErrUserEmailAlreadyExists):
			writeErrorJSON(w, http.StatusConflict, err.Error())
		default:
			// TODO: サーバ側ログに err を記録（slog 等を後で追加）
			writeErrorJSON(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// ---- 4. 成功レスポンス -------------------------------------------
	writeJSON(w, http.StatusCreated, newCreateUserResponse(created))
}

// ============================================================================
// JSON レスポンス用の小さなヘルパ群。
//
// 用途: `c.JSON(status, body)` 相当の処理を自前で行う。
// Gin がやっていた 3 つの操作:
//   1. Content-Type ヘッダを "application/json" にセット
//   2. ステータスコードを書く
//   3. Body を JSON エンコードして流す
// を素直に並べたのがこの writeJSON。
//
// package-private（小文字始まり）なので外部パッケージからは呼べない。
// 将来 category / task ハンドラでも同じ処理が要るので、
// 共通化したくなった時点で internal/handler/response/ 等に切り出す予定。
// ============================================================================

func writeJSON(w http.ResponseWriter, status int, body any) {
	// Header は WriteHeader より前にセットする必要がある。
	// WriteHeader を呼んだ後にヘッダを変えても反映されない（Go の仕様）。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// Encode エラーは握り潰している（client が切断した等で Write 失敗は起こり得る）。
	// 本番ではログに記録するのが望ましい。
	_ = json.NewEncoder(w).Encode(body)
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
