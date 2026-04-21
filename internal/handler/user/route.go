package user

import (
	"github.com/gin-gonic/gin"

	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// RegisterRoutes は /users 配下のルートを登録する。
//
// 引数:
//   - rg         : 親の RouterGroup（例: /api/v1 の下に /users を生やす想定）
//   - createUser : POST /users で呼ぶ usecase
//     ※ 将来 GetUser / UpdateUser 等を足す時は引数を追加する
//
// 設計の意図:
//   - ルートの宣言は**このパッケージに閉じ込める**。main.go で直に gin.POST(...)
//     を並べると、main.go が肥大化し、user のルートが他の場所からも登録できてしまう。
//   - 逆に「ハンドラを外から呼ぶ」機会は想定していないので、エントリポイントを
//     1 つ（この関数）に絞る。
//
// なぜ *gin.RouterGroup を受け取るのか（*gin.Engine ではなく）:
//   - エンジン全体ではなく、/api/v1 のような「プレフィックス付きグループ」配下に
//     ぶら下げたいことが多い
//   - 親側が自由にグループを切れる（ミドルウェアも親側で付けられる）
func RegisterRoutes(rg *gin.RouterGroup, createUser *userusecase.CreateUser) {
	// ここでハンドラ構造体をインスタンス化する。
	// ハンドラ構造体自体は外に露出しない（NewCreateHandler は同パッケージ内で使うだけ）。
	createHandler := NewCreateHandler(createUser)

	// `/users` プレフィックスを一度だけ書くため Group でくくる。
	// POST/GET/... を追加する時に rg.POST("/users/xxx", ...) と毎回書かずに済む。
	users := rg.Group("/users")
	{
		users.POST("", createHandler.Handle)
		// TODO: users.GET("/:userId", getHandler.Handle)  // Step で追加予定
	}
}
