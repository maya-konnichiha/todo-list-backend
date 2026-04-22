# 技術選定と環境構築

本プロジェクト `todo-list-backend` の **技術選定** と **環境構築** の流れを、判断理由とともに記録する。

---

## 1. プロジェクト概要

Go + PostgreSQL で TODO リストの REST API バックエンドを構築する **学習用プロジェクト**。フレームワークや高レベルツールに依存せず、標準ライブラリ中心に 4 層クリーンアーキテクチャを実装する。

- 言語: Go 1.26
- データベース: PostgreSQL 17
- アーキテクチャ: 4 層クリーンアーキテクチャ (`handler` / `usecase` / `domain` / `repository`)
- 参照実装: `material-hub-backend` (設計パターンのみ参考、ライブラリは移植しない)

---

## 2. 技術選定

### 2.1 基本方針

**「フレームワーク・高レベルツールは使わない」** を明確な制約とする。

| 目的 | 期待する学習効果 |
|------|------------------|
| 標準ライブラリで何ができるかを理解する | `net/http`・`encoding/json`・`log/slog` 等の本来の機能を体得 |
| フレームワークの魔法を先に覚えない | 基礎を固め、必要になったときに自分で選定できる |
| SQL を手書きで書く | クエリ最適化やインデックス設計を意識する習慣が付く |

### 2.2 採用技術

| 用途 | 技術 | 選定理由 |
|------|------|----------|
| 言語 | **Go 1.26** | 学習対象 |
| DB | **PostgreSQL 17** | 実務でよく使われる、型が豊富 |
| DB driver | **`github.com/jackc/pgx/v5`** + `pgxpool` | Go における実質的な標準。ドライバなので「ツール」には該当しない |
| マイグレーション | **`golang-migrate`** (CLI、mise の `gomigrate` 経由) | シンプル、SQL ファイル手書き |
| 環境変数 | **`godotenv`** + **`mise`** | `.env` 読み込み |
| HTTP ルーティング | **標準 `net/http` + `http.ServeMux`** | Go 1.22+ のパターン機能 (`"POST /users"`) で REST 可能 |
| JSON | 標準 `encoding/json` | 標準で十分 |
| ログ | 標準 `log/slog` | Go 1.21+ 標準の構造化ロガー |
| メール検証 | 標準 `net/mail` | `ParseAddress` で形式チェック |
| コンテナ | **Docker Compose** | DB のみコンテナ化 |
| ツールバージョン管理 | **`mise`** | Go / golang-migrate を固定 |

### 2.3 採用しない技術(参照実装 material-hub には入っているが不採用)

| 候補 | 代替 | 理由 |
|------|------|------|
| Gin / echo | 標準 `net/http` | Go 1.22+ のパターンルーティングで REST 可能 |
| sqlc / gorm | SQL 手書き | SQL を学ぶため、生成コードに頼らない |
| atlas | golang-migrate | 学習用には高機能すぎる |
| mockery | (現時点では未採用) | テスト整備時に再検討 |
| `go-playground/validator` | 手書きバリデーション | `net/mail` + 文字数カウントで十分 |

---

## 3. アーキテクチャ

### 3.1 4 層クリーンアーキテクチャ

```
internal/
├── domain/           # エンティティ、Repository インターフェース、ドメインエラー
│   └── user/
├── repository/       # domain のインターフェースを満たす DB 実装
│   └── user/
├── usecase/          # ビジネスロジック
│   └── user/
└── handler/          # HTTP ハンドラ + ルーティング + エラー変換
    ├── errhandler/   # エラー → JSON レスポンス変換
    ├── user/         # エンドポイント別ハンドラ
    └── router.go     # ルーティング設定
```

### 3.2 依存方向

```
handler ──▶ usecase ──▶ domain ◀── repository
                           (interface)    (実装)
```

- **domain** は何にも依存しない(標準 `context` / `errors` / `time` のみ)
- **repository** は domain のインターフェースを満たす具象実装を提供
- **usecase** は domain のインターフェース経由で repository を使う(具象を知らない)
- **handler** は usecase を呼び、入出力を担当

この向きで依存させることで、「DB 実装を差し替えても usecase/domain は無修正」「usecase を単体テストするときに mock repository を差し込める」状態になる。

---

## 4. 環境構築

### 4.1 前提ツール

| ツール | 用途 |
|--------|------|
| [mise](https://mise.jdx.dev/) | Go / golang-migrate のバージョン管理 |
| [Docker Desktop](https://www.docker.com/) | PostgreSQL コンテナ実行 |
| [gh CLI](https://cli.github.com/) | GitHub リポジトリ操作(任意) |

### 4.2 ディレクトリ構成

```
todo-list-backend/
├── cmd/server/              # エントリーポイント (main.go で DI 配線)
├── internal/                # アプリ内部コード
│   ├── domain/
│   ├── repository/
│   ├── usecase/
│   └── handler/
├── migrations/              # *.up.sql (forward-only)
├── docs/                    # ドキュメント(このファイルも含む)
├── .env.example             # 環境変数テンプレート
├── .env                     # 実際の環境変数 (.gitignore 対象)
├── docker-compose.yml       # Postgres サービス定義
├── mise.toml                # ツールバージョン + タスク定義
├── go.mod                   # Go モジュール
└── README.md
```

### 4.3 セットアップ手順

```sh
# 1. 環境変数ファイル作成
cp .env.example .env

# 2. mise でツールインストール (Go + gomigrate)
mise install

# 3. 依存パッケージ取得
mise run tidy

# 4. Postgres コンテナ起動
mise run db-up

# 5. マイグレーション適用
mise run migrate-up

# 6. サーバー起動
mise run run
```

### 4.4 mise タスク一覧

| タスク | 説明 |
|--------|------|
| `mise run db-up` | Postgres コンテナを起動 |
| `mise run db-down` | Postgres コンテナを停止 |
| `mise run db-logs` | Postgres のログを追従 |
| `mise run db-shell` | `psql` で DB に入る |
| `mise run migrate-up` | マイグレーション全適用 |
| `mise run migrate-create <name>` | 新規マイグレーションファイル生成 |
| `mise run run` | サーバー起動 |
| `mise run tidy` | `go mod tidy` |

### 4.5 ポート衝突への対処

ホスト側に既存の PostgreSQL が稼働していると、デフォルトの 5432 は使えない(`Error response from daemon: ports are not available: listen tcp 0.0.0.0:5432: bind: address already in use`)。

対応方法: `.env` で使用ポートを変更する。

```sh
# .env (例: 9999 を使う場合)
POSTGRES_PORT=9999
DATABASE_URL=postgres://todo_user:todo_password@localhost:9999/todo_db?sslmode=disable
```

**注意**: `DATABASE_URL` 内にもポートがハードコードされているため、`POSTGRES_PORT` と **両方** を揃える必要がある。二重管理が嫌なら将来的に変数展開で一本化する:

```sh
DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
```

### 4.6 Docker 対象の範囲

**DB のみ Docker 化し、Go アプリ本体はホストで `go run` 実行する。**

| 目的 | 対象 | 実施 |
|------|------|------|
| 開発環境の統一(DB バージョン揃える) | 依存サービス(DB など) | ✅ Docker 化 |
| デプロイ成果物の配布 | アプリ本体 | ⛔ 当面実施しない |

material-hub のように app 自体も Docker に入れるのは、本番デプロイ・チーム開発段階で追加する判断。学習段階では `go run` の速い反復ループを優先する。

---

## 5. データベース設計

### 5.1 テーブル一覧

- **users**: ユーザー情報 (`user_id`, `user_name`, `user_email`, timestamps, `deleted_at`)
- **categories**: ユーザー別カテゴリ (`category_id`, `user_id` FK, `category_name`, timestamps, `deleted_at`)
- **tasks**: タスク (`task_id`, `user_id` FK, `category_id` FK nullable, `task_title`, `task_description`, `task_status`, `task_due_at`, timestamps, `deleted_at`)

### 5.2 設計判断の記録

| 項目 | 決定 | 理由 |
|------|------|------|
| 主キー型 | `BIGINT GENERATED ALWAYS AS IDENTITY` | `SERIAL` より推奨、手動セット不可で安全 |
| Soft Delete | 全テーブルに `deleted_at` カラム | 監査・復元を見越す |
| FK `categories.user_id → users` | `ON DELETE CASCADE` | ユーザー物理削除時に紐づくカテゴリも削除 |
| FK `tasks.user_id → users` | `ON DELETE CASCADE` | ユーザー物理削除時に紐づくタスクも削除 |
| FK `tasks.category_id → categories` | `ON DELETE SET NULL` | カテゴリ削除でもタスクは残す |
| Email UNIQUE + Soft Delete | 通常 UNIQUE + 削除時に**アプリ層で email リネーム**(パターン B) | 削除済みユーザーの email 再利用可。現代 SaaS でよく見るパターン |
| `updated_at` 自動更新 | **アプリ層で** `SET updated_at = CURRENT_TIMESTAMP` を明示 | SQL が素直、挙動が可視 |
| インデックス | FK 3 本の部分インデックス (`WHERE deleted_at IS NULL`) | 学習用に最小構成 |
| `task_status` | `VARCHAR(20)` + `CHECK (task_status IN ('todo', 'in_progress', 'done'))` | ENUM より変更容易 |
| `task_due_at` | `TIMESTAMPTZ` | 時刻指定可能、タイムゾーン管理の学習 |
| Triangular FK 対策 | アプリ層で `category.user_id == currentUserID` を検証 | 複合 FK は `ON DELETE SET NULL` と両立しない |

### 5.3 Soft Delete + Email UNIQUE の選択(パターン B)

削除済みユーザーの email を再登録可能にするためのパターンを 4 候補から検討:

| パターン | 内容 | 選定 |
|----------|------|------|
| A: 匿名化 | 削除時に email を `deleted-{id}@anonymized.local` に変換 + 物理削除 or 個人情報消去 | 未採用 |
| **B: リネーム方式** | 削除時に email を `deleted_{timestamp}_{original}` にアプリ層で変更 | **採用** |
| C: 猶予期間 | N 日間保持 → 物理削除を cron で実行 | 未採用 |
| D: 永久保持 | 削除済みでも email は永久予約 | 未採用 |

**B を選んだ理由**: 通常 UNIQUE のままで済み、アプリ層のロジックだけで完結する。Linear や Notion 等の B2B SaaS でもよく見られる。

---

## 6. マイグレーション運用

### 6.1 forward-only 方針

- `.up.sql` のみ使用、`.down.sql` は **作らない**
- `.up.sql` 冒頭に `DROP TABLE IF EXISTS <table>;` を入れて自己完結型にする
- material-hub / atlas と同じ思想
- ロールバックは volume ごと削除する運用に割り切る

### 6.2 ファイル構成

```
migrations/
├── 000001_create_users.up.sql
├── 000002_create_categories.up.sql
└── 000003_create_tasks.up.sql
```

各ファイルの構造(例: `000001_create_users.up.sql`):

```sql
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    user_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_name VARCHAR(50) NOT NULL,
    user_email VARCHAR(255) NOT NULL UNIQUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 6.3 リセット手順(「rollback」の代替)

```sh
docker compose down -v   # コンテナ + volume を削除
mise run db-up           # 再起動
mise run migrate-up      # 再適用
```

### 6.4 新規マイグレーションの追加

```sh
mise run migrate-create add_something
# → migrations/000004_add_something.up.sql が生成される
# ファイル冒頭に `DROP TABLE IF EXISTS ...;` を追加してから編集
```

---

## 7. Git リポジトリ

### 7.1 リモート切り替え履歴

初期 clone 元: `github.com/maya-konnichiha/todo-list-backend`(他人のリポジトリ)
↓
自分のリポジトリへ: `github.com/Izushi/todo-list-backend`(Public)

### 7.2 実行コマンド

```sh
# 1. GitHub 上に新規リポジトリ作成
gh repo create Izushi/todo-list-backend --public --description "Go + PostgreSQL TODO list backend (learning project)"

# 2. ローカルの origin URL を変更
git remote set-url origin git@github.com:Izushi/todo-list-backend.git

# 3. push + upstream 設定
git push -u origin main
```

---

## 8. 次のステップ

- [x] 技術選定
- [x] 環境構築(Docker + mise + migrations)
- [x] DB 設計 + マイグレーション適用
- [x] User 作成 API (`POST /users`) 実装
- [ ] User 取得 API (`GET /users/{userId}`) 実装 ← 次
- [ ] X-User-ID 認証ミドルウェア
- [ ] Category / Task API
- [ ] ユニットテスト / 統合テスト
