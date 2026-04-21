package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	// 自分のパッケージ名も user なので、衝突回避のため alias を付ける。
	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// CreateHandler は POST /users を処理するハンドラ。
//
// 構造体にして usecase を保持するのは usecase 層と同じ DI の理由:
//   - 依存（ここでは CreateUser ユースケース）を外から注入する
//   - テストでモック usecase に差し替えられる
//   - ハンドラ関数を登録するたびに依存を並べる必要がなくなる
type CreateHandler struct {
	usecase *userusecase.CreateUser
}

// NewCreateHandler はコンストラクタ。
//   - main.go で配線する時に使う
//   - *userusecase.CreateUser を受け取り、自身のフィールドに保持する
func NewCreateHandler(uc *userusecase.CreateUser) *CreateHandler {
	return &CreateHandler{usecase: uc}
}

// Handle は Gin の HandlerFunc。`func(c *gin.Context)` のシグネチャに合う。
//
// 処理の 4 ステップ:
//  1. リクエスト JSON を request DTO にパース & バリデーション
//  2. usecase に渡す params に詰め替えて呼び出し（context は Gin から取得）
//  3. usecase のエラーを HTTP ステータスコードに変換
//  4. 成功時は response DTO に詰め替えて 201 Created で返却
func (h *CreateHandler) Handle(c *gin.Context) {
	// ---- 1. パース & バリデーション ----------------------------------
	//
	// c.ShouldBindJSON:
	//   - リクエスト Body の JSON を構造体にデコード
	//   - binding タグ（required, email, max=50 等）を検証
	//   - 失敗時はエラーを返す（ここで 400 を返す）
	//
	// `c.Bind...` 系は失敗時に自動で 400 を書き込むが、エラーメッセージを
	// 自分で制御したいので `ShouldBind...` を使う（"Should" = 自動返信しない版）。
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// err.Error() には go-playground/validator の詳細メッセージが入る。
		// 学習段階ではそのまま見せるが、プロダクションでは内部情報を隠す/
		// 構造化する（例: {"errors": [{"field": "user_name", "reason": "required"}]}）等が望ましい。
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ---- 2. usecase 呼び出し ----------------------------------------
	//
	// context の取り方:
	//   - `c.Request.Context()` で Gin が保持しているリクエストスコープの ctx を取得
	//   - この ctx を usecase → repository → pgx へと素通しする
	//   - クライアントが接続を切れば ctx.Done() が閉じ、全下流のクエリも中断される
	//
	// なぜ handler の c（gin.Context）を usecase に渡さないのか:
	//   - gin.Context は HTTP 固有の機能（ヘッダ取得、リクエストパースなど）を持つ
	//   - usecase が gin.Context を受け取ると Gin 依存になってしまい、
	//     HTTP 以外から呼び出せなくなる（層の境界が崩れる）
	//   - なので「context.Context（標準）」だけを渡す
	created, err := h.usecase.Execute(c.Request.Context(), userusecase.CreateUserParams{
		UserName:  req.UserName,
		UserEmail: req.UserEmail,
	})
	if err != nil {
		// ---- 3. エラーを HTTP ステータスに変換 ----------------------
		//
		// sentinel エラー → HTTP ステータスの対応:
		//   400 Bad Request : 業務入力がドメインルールに反する
		//   409 Conflict    : 他のリソースと競合（email 重複）
		//   500 Internal    : 想定外（DB 接続失敗等）
		//
		// errors.Is は「エラーチェーンの中に指定の sentinel が含まれるか」を確認する。
		// %w でラップされた多段エラーでも透過的に辿れる（%v ラップは辿れない）。
		switch {
		case errors.Is(err, userdomain.ErrUserNameEmpty),
			errors.Is(err, userdomain.ErrUserNameTooLong),
			errors.Is(err, userdomain.ErrUserEmailEmpty),
			errors.Is(err, userdomain.ErrUserEmailInvalid):
			// ドメインルール違反 = クライアントの入力が悪い = 400
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		case errors.Is(err, userdomain.ErrUserEmailAlreadyExists):
			// email の UNIQUE 衝突 = リソース競合 = 409
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

		default:
			// それ以外は想定外。内部エラーメッセージをそのまま返すと
			// 実装詳細（SQL エラー、スタック等）を漏らす危険があるので、
			// クライアントには汎用的な文言だけ返す。
			// 実際のエラーはサーバ側ログに記録する（今はまだ logger を入れていないので TODO）。
			// TODO: ログ出力（slog 等）を後で追加する
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// ---- 4. 成功時のレスポンス --------------------------------------
	//
	// ステータス 201 Created:
	//   - 「リソースが新規作成された」を表す HTTP 標準のコード
	//   - 200 OK より意味が明確
	//
	// domain.User をそのまま JSON 化せず、newCreateUserResponse で DTO に詰め替える。
	// response.go の冒頭コメントで理由を詳述。
	c.JSON(http.StatusCreated, newCreateUserResponse(created))
}
