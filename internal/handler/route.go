// Package handler は HTTP 層の**束ね役**。
//
// このファイル（handler/route.go）は以下を担当する:
//   - Gin エンジン全体へのルート登録（ヘルスチェック等のトップレベル）
//   - /api/v1 グループを作り、各サブパッケージ（user, task, category, ...）の
//     RegisterRoutes を呼び出してルート登録を委譲する
//
// ハンドラ層が「2 段構造」になっている理由:
//
//  1. 関心の分離:
//     - handler/user/route.go は "user のルート"（POST /users, GET /users/:id）だけを知る
//     - handler/route.go は "全体の組み立て方"（/api/v1 配下に user を足す）を知る
//     両者が混ざると、エンティティ追加のたびに中央が肥大化する。
//
//  2. 単一のエントリポイント:
//     cmd/server/main.go からは handler.RegisterRoutes(engine, deps) 一本を呼べば
//     完結する。エンティティが増えても main.go のコードは**変わらない**。
//
//  3. 拡張性:
//     新しいエンティティを足す時は
//       a. handler/{entity}/route.go に RegisterRoutes を作る
//       b. Deps に usecase を 1 行追加
//       c. v1 グループ配下に 1 行追加
//     だけで済む。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	// サブパッケージの handler をまとめて import する。
	// 自分のパッケージ名は `handler` なので、`user` という名前と衝突しないが、
	// usecase 側の user と区別しやすいよう alias を付ける（`userhandler`）。
	userhandler "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"

	// usecase 側も user パッケージ。handler 側 user と名前衝突するので alias 必須。
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// Deps は全ハンドラが必要とする依存（usecase 等）をまとめた構造体。
//
// なぜ引数を個別に並べず Deps 構造体にするのか:
//   - エンティティが増えるたびに RegisterRoutes のシグネチャが伸びるのを防ぐ
//     （関数引数が 20 個並ぶ main.go は読めない）
//   - フィールド名が呼び出し側に見えるので、どの usecase をどこに渡したか明確
//   - 並び順の事故（別 usecase を取り違える）が起きにくい
//   - 将来 Logger や Config を追加する時も、フィールド 1 行足すだけで済む
//
// なぜ usecase を Deps で**外から受け取る**のか（DI の続き）:
//   - handler は usecase を自分で new しない → 依存方向を一方向に保てる
//   - テスト時にモック usecase を渡せる
//   - usecase のライフサイクル（いつ作る / 破棄するか）を main に集約できる
//
// 将来の拡張例:
//
//	type Deps struct {
//	    CreateUser *userusecase.CreateUser
//	    GetUser    *userusecase.GetUser
//
//	    CreateCategory *categoryusecase.CreateCategory
//	    ListCategories *categoryusecase.ListCategories
//	    ...
//
//	    Logger *slog.Logger
//	}
type Deps struct {
	// User 関連のユースケース
	CreateUser *userusecase.CreateUser
	// TODO: GetUser / UpdateUser 等を追加

	// TODO: Category / Task のユースケースも同様に追加
}

// RegisterRoutes は Gin エンジン全体にルートを登録する。
//
// 引数:
//   - r : gin.Engine。アプリ全体のルーターインスタンス。
//   - d : Deps（usecase 群）
//
// gin.Engine と gin.RouterGroup の違い:
//
//   - gin.Engine: アプリのルート（最上位）。`gin.Default()` や `gin.New()` で作るトップレベルのもの。
//     ミドルウェア登録や全体の設定（Trust Proxy 等）もここで行う。
//     `engine.Run(":8888")` でサーバー起動もここから。
//
//   - gin.RouterGroup: 「共通プレフィックスを持つルート集合」。
//     engine 自体も実は RouterGroup の埋め込みなので、engine.GET(...) も
//     動くが、共通プレフィックス ("/api/v1") や共通ミドルウェアを効かせたい
//     スコープでは `engine.Group("/api/v1")` のように **子グループ** を切る。
//
//   - サブパッケージ（handler/user 等）の RegisterRoutes は *gin.RouterGroup を
//     受け取る設計にしている。これによりサブパッケージは「自分がどのプレフィックス配下に
//     置かれるか」を知らなくて良い（置く場所は親が決める）。
func RegisterRoutes(r *gin.Engine, d Deps) {
	// --- トップレベル: ヘルスチェック -----------------------------------
	//
	// /api/v1 の外に置くのは、インフラ（ロードバランサやモニタリング）が叩く用途を想定しているため。
	// API のバージョンに左右されないパスにしておく方がインフラ側から安定して参照できる。
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --- /api/v1 グループ -----------------------------------------------
	//
	// なぜ /api/v1 という 2 階層のプレフィックス？
	//   - /api     : 「API エンドポイント」であることを明示（将来 SSR ページ等と共存しても衝突しない）
	//   - /v1      : バージョン。破壊的変更（API 形式が変わる）時は /v2 を並列に走らせて、
	//                旧クライアントと新クライアントを同居させられる。
	//   → フロントやモバイルとの互換性を**段階的に**切り替えたい時に効く。
	//
	// ここで切った v1 グループを各サブパッケージに渡すことで、
	// 各パッケージは "/api/v1" という文字列を知らずに済む
	// （=ルートプレフィックスの変更時、サブパッケージを触らずこのファイルだけで済む）。
	v1 := r.Group("/api/v1")
	{
		// User 関連のルートを登録
		// サブパッケージ側で POST /users 等を v1 配下にぶら下げる。
		userhandler.RegisterRoutes(v1, d.CreateUser)

		// TODO: Category / Task の追加時はここに 1 行ずつ足す
		//   categoryhandler.RegisterRoutes(v1, d.CreateCategory, d.ListCategories, ...)
		//   taskhandler.RegisterRoutes(v1, d.CreateTask, d.ListTasks, ...)
	}
}
