package user

import (
	userdomain "github.com/maya-konnichiha/todo-list-backend/internal/domain/user"
)

// CreateUserResponse は POST /users のレスポンス。
//
// フロントの要件:
//   - 作成直後に「登録できました、〇〇さん」と名前を表示したい
//   - 追加の GET を不要にするため 1 回で必要情報を返す
// → user_id / user_name / user_email の 3 つだけ返す。
//
// JSON タグ（`json:"user_id"`）は snake_case に揃える（リクエスト側と統一）。
type CreateUserResponse struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// newCreateUserResponse は domain の User をレスポンス DTO に詰め替えるヘルパ。
//
// なぜ domain.User をそのまま JSON 化して返さないのか:
//  1. **外部公開したくないフィールドを隠せる**
//     User は CreatedAt / UpdatedAt / DeletedAt を保持しているが、
//     今回の API ではレスポンスに含めない。DTO に詰め替えることで「公開したい
//     ものだけ露出」が明示的になる。
//  2. **API の形を domain から独立させられる**
//     domain のフィールド名や型を変えても、DTO の JSON 形は保てる（API 互換性）。
//     逆に JSON 形を変えたくても、domain を触らずに DTO だけ調整できる。
//  3. **型変換の余地を持てる**
//     例: UserID を int64 → "12345"（string）に変えたい API 仕様でも、ここで変換できる。
//     domain は数値のまま、API は string、と分けられる。
//
// 関数名が小文字始まり（newCreateUserResponse）= パッケージ外から呼ばれない。
// response 変換はこのパッケージ内部の詳細なので、export する必要がない。
func newCreateUserResponse(u *userdomain.User) CreateUserResponse {
	return CreateUserResponse{
		UserID:    u.UserID,
		UserName:  u.UserName,
		UserEmail: u.UserEmail,
	}
}
