// Package handler は HTTP 層の束ね役。
//
// このファイル（handler/route.go）は:
//   - ServeMux 全体へのルート登録（ヘルスチェック等のトップレベル）
//   - /api/v1 をプレフィックスとして各サブパッケージ（user, task, category, ...）の
//     RegisterRoutes を呼び出し、ルート登録を委譲する
//
// 2 段構造のねらいや拡張方針は前回の設計と同じ:
//  1. 関心の分離: user 側は自分のルートだけ知る
//  2. main.go のエントリ一点化: handler.RegisterRoutes(mux, deps) 1 本
//  3. 拡張性: エンティティ追加時は handler/{entity}/route.go + Deps 1 行 + 1 行呼び出し
package handler

import (
	"encoding/json"
	"net/http"

	userhandler "github.com/maya-konnichiha/todo-list-backend/internal/handler/user"
	userusecase "github.com/maya-konnichiha/todo-list-backend/internal/usecase/user"
)

// Deps は全ハンドラが必要とする依存（usecase 等）をまとめた構造体。
// フィールドを増やすだけでエンティティを追加できる拡張ポイント。
type Deps struct {
	// User 関連のユースケース
	CreateUser *userusecase.CreateUser

	// TODO: GetUser / UpdateUser / Category / Task のユースケースをここに追加
}

// RegisterRoutes は ServeMux にアプリ全体のルートを登録する。
//
// 引数:
//   - mux : http.ServeMux（cmd/server/main.go で `http.NewServeMux()` で作る）
//   - d   : Deps（usecase 群）
//
// Gin 時代との対比:
//
//	Gin:  RegisterRoutes(r *gin.Engine, d Deps)   — Engine は RouterGroup の埋め込み
//	std:  RegisterRoutes(mux *http.ServeMux, d Deps) — Engine 相当が ServeMux
//
// ServeMux の特徴:
//   - Go 1.22+ で "METHOD /path/{param}" 形式に対応し、Gin 相当の表現力になった
//   - "Group" 概念は無いので、パスプレフィックスは**文字列**でサブパッケージに渡す
//   - パニックリカバリーやリクエストログは**自動では付かない**。必要なら自前でミドルウェア実装
//     （Gin の gin.Default() が入れていたもの）
func RegisterRoutes(mux *http.ServeMux, d Deps) {
	// --- トップレベル: ヘルスチェック -----------------------------------
	// バージョンに左右されない /health に置き、LB / モニタ用途を想定。
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// --- /api/v1 配下 ---------------------------------------------------
	//
	// Gin の r.Group("/api/v1") は内部的にプレフィックスを保持した RouterGroup を返すが、
	// ServeMux には同等の仕組みが無い。代わりに「プレフィックス文字列」を
	// サブパッケージの RegisterRoutes に渡して、先頭に付けさせる。
	//
	// この設計の副産物:
	//   - サブパッケージは "/api/v1" を知らない。変えたければこのファイルだけ触れば良い
	//   - "/api/v1" と "/internal-admin" 等、複数プレフィックス配下に同じ user routes を
	//     生やすのも簡単になる（呼び出し側で prefix を変えて 2 回呼ぶだけ）
	const v1 = "/api/v1"
	userhandler.RegisterRoutes(mux, v1, d.CreateUser)

	// TODO: categoryhandler.RegisterRoutes(mux, v1, d.CreateCategory, ...)
	// TODO: taskhandler.RegisterRoutes(mux, v1, d.CreateTask, ...)
}
