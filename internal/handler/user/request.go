// Package user は /users エンドポイント群の HTTP ハンドラを提供する。
//
// 他の層との切り分け:
//   - domain:  「User とは」のビジネスルール & バリデーション
//   - usecase: 「User を作る手順」
//   - handler: JSON をデコードし、usecase を呼び出し、結果を JSON で返す
//
// ディレクトリが internal/handler/user/ なのでパッケージ名は `user`。
// domain/user・usecase/user と名前衝突するため、import では alias を付ける。
package user

// CreateUserRequest は POST /users のリクエストボディ。
//
// フィールドバリデーションは**ここでは行わない**。
//
// なぜ HTTP 層でバリデーションしないか:
//   - 入力の正否判定は `domain/user.NewUser` が担う（sentinel エラーで通知）
//   - 二重にルールを書くと、ルールが食い違った時にバグの温床になる
//   - 「正しさの真実は domain にだけ存在する」と統一できる
//
// HTTP 層で行うのは**構文レベルのチェックだけ**:
//   - Body が JSON として壊れていないか（json.Decode が拾う）
//   - フィールドの型があっているか（string のところに数値が来たら Decode がエラー）
//
// JSON タグ（`json:"user_name"`）は変換キー名なので必須。これは API 仕様の一部。
type CreateUserRequest struct {
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}
