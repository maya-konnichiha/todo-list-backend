// Package user は /users エンドポイント群の HTTP ハンドラを提供する。
//
// 他の層との切り分け:
//   - domain:  「User とは」のビジネスルール
//   - usecase: 「User を作る手順」
//   - handler: 「HTTP の JSON を usecase 入力に変換し、結果を JSON に返す」
//     ここ（handler）は HTTP 固有の知識（JSON、ステータスコード、ヘッダ）だけを扱う。
//
// ディレクトリが internal/handler/user/ なのでパッケージ名は `user`。
// domain/user・usecase/user と名前衝突するため、import では alias を付ける。
package user

// CreateUserRequest は POST /users のリクエストボディ。
//
// 「リクエスト DTO」を usecase の入力構造体（CreateUserParams）と**別に**定義するのは、
//   - API の JSON 形が変わっても usecase/domain が揺れないようにするため
//   - usecase を HTTP 以外（CLI / ワーカー等）から呼ぶ場面で DTO を経由しないため
//
// タグの読み方:
//
//	`json:"user_name"`                 ← JSON キー名
//	`binding:"required,max=50"`        ← Gin がバインディング時に検証する
//
// 各 binding の意味:
//   - required … フィールドが存在しゼロ値でないこと（string なら空文字でないこと）
//   - max=50   … 文字数の上限（go-playground/validator は rune 数で数える）
//   - email    … RFC 5322 ベースの形式チェック
//
// なぜ DB と同じ上限を書くのか:
//   - handler で早期に 400 を返せる（DB まで行かずに済む = 負荷軽減 + 明快なエラー）
//   - 最終的な「業務上正しいか」は domain の NewUser が、最後の砦として再検証する
//     → **handler バリデーションは最適化、domain バリデーションは正義**
type CreateUserRequest struct {
	UserName  string `json:"user_name"  binding:"required,max=50"`
	UserEmail string `json:"user_email" binding:"required,email,max=255"`
}
