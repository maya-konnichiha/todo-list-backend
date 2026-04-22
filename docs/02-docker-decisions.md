# Docker に関する意思決定

本プロジェクト `todo-list-backend` における Docker 関連の設計判断を記録する。

**扱うトピック:**
- なぜ DB (Postgres) を Docker 化したのか
- なぜ Go アプリを Docker 化しないのか
- なぜ `Dockerfile` が存在しないのか

---

## 1. 現状の Docker 構成

```
┌───────────────────────────────────────────────┐
│ ホスト (Mac / Linux)                          │
│                                               │
│  ┌─────────────────────────────────────────┐  │
│  │ Go アプリ                               │  │
│  │   $ mise run run                        │  │
│  │   $ go run ./cmd/server                 │  │
│  │   (ホスト上で直接実行)                   │  │
│  └─────────────────────────────────────────┘  │
│               ↓ localhost:9999 で接続         │
│  ┌─────────────────────────────────────────┐  │
│  │ Docker Desktop                          │  │
│  │                                         │  │
│  │   ┌───────────────────────────────┐     │  │
│  │   │ container: todo-postgres      │     │  │
│  │   │   image: postgres:17-alpine   │     │  │
│  │   │   port: 5432 (内部)           │     │  │
│  │   └───────────────────────────────┘     │  │
│  └─────────────────────────────────────────┘  │
└───────────────────────────────────────────────┘
```

| 構成要素 | 実行場所 | image / ファイル |
|----------|----------|------------------|
| Go アプリ | **ホスト** | なし(`go run` で直接実行) |
| PostgreSQL | **Docker コンテナ** | `postgres:17-alpine`(公式 image) |
| `docker-compose.yml` | プロジェクトルート | Postgres の起動設定 |
| `Dockerfile` | **存在しない** | 自前ビルド不要のため |

---

## 2. DB (Postgres) を Docker 化した理由

個人開発・学習用途であっても、Postgres のような **外部ミドルウェア** は Docker 化する価値が大きい。

### 2.1 インストールが煩雑

**ネイティブインストールの手段が複数ある:**
- Homebrew (`brew install postgresql@17` + launchctl 設定)
- Postgres.app (GUI)
- 公式インストーラ
- Docker 以外で入れると、これらの混在・競合・アンインストール残留を招く

**Docker なら:** `compose.yml` に 1 行書くだけで完了。ホストを一切汚さない。

### 2.2 プロジェクトごとに異なるバージョンを使える

- プロジェクト A: Postgres 15
- プロジェクト B: Postgres 17
- プロジェクト C: Postgres 16

ホスト側に複数バージョンを同居させるのは煩雑。Docker なら `compose.yml` に `image: postgres:15` / `:17` と書き分けるだけ。**環境干渉ゼロ**。

### 2.3 捨てやすい → 実験が怖くなくなる

学習では **「DB を壊して作り直す」** を何度もやる:
- マイグレーションの試行錯誤
- スキーマ変更の練習
- テストデータを何度も入れ直す

```sh
docker compose down -v   # DB の volume ごと削除
docker compose up -d     # 新品の DB で再起動
mise run migrate-up
```
**20 秒で完全リセット** できる。ネイティブ Postgres では「データベースを物理削除」を気軽にできない心理的抵抗が壊れる。

### 2.4 バックグラウンドで常駐しない

| | ネイティブ Postgres | Docker Postgres |
|--|---------------------|------------------|
| Mac 起動時の挙動 | 自動起動(brew services) | 起動しない |
| 使わない時のリソース | メモリを占有し続ける | 0(停止しておける) |
| バッテリー消費 | 常時影響 | 使う時のみ |

`docker compose stop` で止めておけるので、使わない時のリソース消費が少ない。

### 2.5 設定が `docker-compose.yml` に集約 → 自己文書化

```yaml
postgres:
  image: postgres:17-alpine
  ports: ["9999:5432"]
  environment:
    POSTGRES_USER: todo_user
    POSTGRES_PASSWORD: todo_password
    POSTGRES_DB: todo_db
  volumes:
    - postgres_data:/var/lib/postgresql/data
```

このファイルだけ見れば:
- どのバージョンの Postgres を使っているか
- ポート、ユーザー名、DB 名
- データの永続化方式

**すべてがコードとして読める**。README に手順を書き連ねる必要がない。

### 2.6 マシン移行・他人への共有が容易

- 新しい Mac へ移行: `git clone && docker compose up -d` だけで同じ環境が復元
- 他人と共有: 「Docker 入れて、compose up して」の 1 ステップ
- Postgres 固有の設定(pg_hba.conf、postgresql.conf 等)を個別に教える必要がない

### 2.7 本番 (Linux コンテナ) との環境一致

- 本番 Postgres は Linux で動く
- ローカルも Linux コンテナで動かせば、タイムゾーン・照合順序・FS 大文字小文字扱い等の差分が消える
- 「自分の Mac では動いたのに本番で挙動が違う」事故を予防

### 2.8 まとめ表

| メリット | 個人開発での効き |
|----------|-------------------|
| インストール地獄の回避 | ✅ 強 |
| 複数バージョン共存 | ✅ 強 |
| 完全リセットが容易 | ✅ 強(特に学習で効く) |
| ホストを汚さない | ✅ 強 |
| 常駐しない | ✅ 中 |
| 設定の自己文書化 | ✅ 中 |
| マシン移行が楽 | ✅ 中 |
| 本番との環境一致 | ⚠️ 将来デプロイするなら |

---

## 3. Go アプリを Docker 化しない理由

**Docker 化は無条件に良いわけではなく、対象によって効きが違う。** 個人開発の Go アプリでは恩恵が小さく、コストが勝る。

### 3.1 Postgres と Go アプリの違い

| 観点 | Postgres | Go アプリ |
|------|----------|-----------|
| インストールが重い | ✅ | ❌ `mise install` で 1 秒 |
| 複数バージョン同居が面倒 | ✅ | ❌ mise で切替可能 |
| 常駐 / バックグラウンド動作 | ✅ | ❌ 起動しているときだけ動く |
| 状態(データ)を持つ | ✅ | ❌ ステートレス |
| OS ネイティブ依存が深い | ✅ | ❌ Go 静的バイナリは依存ほぼなし |

→ **Postgres は Docker が解決する問題が多い。Go アプリはそもそも問題が少ない。**

### 3.2 Docker 化した場合のコスト

#### ① 反復速度が落ちる
- **ホスト `go run`**: 1-2 秒
- **Docker 内 `go run`**(rerun 経由): 3-5 秒 + Mac の bind mount オーバーヘッド
- 学習 = 試行錯誤の回数 = 速度が命

#### ② デバッガ連携が面倒
- **ホスト**: IDE のブレークポイントがそのまま効く
- **Docker**: `delve` をコンテナ内で起動 → ポート公開 → IDE からリモートアタッチ設定が必要
- 学習初期の壁が増える

#### ③ IDE / LSP / 補完の体験が下がる
- `gopls` / `golangci-lint` / `goimports` はホスト側の Go を見るのが最速
- Docker 越しの LSP は設定が複雑、動作も遅い
- コード補完が効かないと学習ストレスが増える

#### ④ 学習コストが二重になる
- 「Go + クリーンアーキテクチャ」を教えるのに、同時に「Docker + bind mount + rerun」も教える必要が出る
- メンティーのバッファが溢れる

#### ⑤ Docker Desktop のリソース消費
- Mac の Docker Desktop は常時 1-2 GB のメモリ + CPU を占有
- コンテナ数が増えるほど負担増

#### ⑥ Mac の bind mount は遅い
- Mac + Docker の bind mount は FS 同期オーバーヘッドがある
- ファイル監視(rerun 等)の遅延が発生しやすい

### 3.3 Docker 化のメリットが発動しにくい

Postgres で効いた理由を Go アプリに当てはめても、ほぼ刺さらない:

| Postgres で効いた理由 | Go アプリでは… |
|-----------------------|------------------|
| インストールが地獄 | ❌ `mise install` で一瞬 |
| 複数バージョン同居 | ❌ mise.toml で切替可能 |
| バックグラウンド常駐が邪魔 | ❌ Go アプリは常駐しない |
| データの永続化・捨てやすさ | ❌ アプリはステートレス |
| ホストを汚さない | ⚠️ そもそも Go バイナリは汚さない |

### 3.4 Go アプリも Docker 化した方が良い場面(本プロジェクトは該当しない)

| 条件 | 該当? |
|------|-------|
| 複数マイクロサービスを同時起動する | ❌ app + DB だけ |
| デプロイを想定している | ❌ 学習用 |
| sqlc / swag / mockery など多くの周辺ツールを使う | ❌ 標準ライブラリ中心 |
| Redis / Kafka 等 複数の依存サービスと繋ぐ | ❌ Postgres のみ |
| CGO や glibc 依存ライブラリを使う | ❌ 純 Go |
| 他人と共同開発する予定がある | ❌ 1 人学習 |

→ **1 つも当てはまらない → Docker 化のコストのみ、メリットなし** → ホスト実行を選択。

### 3.5 参考: material-hub の場合

material-hub は真逆で、上記の条件にほぼ全て該当するため app も Docker 化している。**プロジェクトの性質によって最適解が変わる** 典型例。

---

## 4. なぜ `Dockerfile` が存在しないのか

### 4.1 Dockerfile の役割(復習)

**Dockerfile = 自前の image を作るためのレシピ。**

| シナリオ | Dockerfile 必要? |
|----------|------------------|
| 自前のアプリを image 化したい | **✅ 必要** |
| 既成 image に改造を加えたい | **✅ 必要** |
| 既成 image を **そのまま** 使う | **❌ 不要** |
| Docker 自体を使わない | **❌ 不要** |

### 4.2 このプロジェクトのサービス別判定

| サービス | Docker 化? | image をどうする? | Dockerfile 必要? |
|----------|------------|---------------------|-------------------|
| PostgreSQL | **Yes** | 公式 `postgres:17-alpine` をそのまま利用 | **❌ 不要** |
| Go アプリ | **No**(ホスト実行) | そもそも image 化していない | **❌ 不要** |

→ **結果、Dockerfile を 1 つも書く必要がない。**

### 4.3 `compose.yml` の該当部分

```yaml
services:
  postgres:
    image: postgres:17-alpine   # ← Docker Hub から pull するだけ
    # build: ... はない
```

`image:` は「Docker Hub などから既成 image を持ってくる」指示。`build: ...` と書くとその場所にある Dockerfile をビルドするが、今回は **`image:` しか書いていない**。

### 4.4 Docker Hub と公式 image

`postgres:17-alpine` の実体は:

```
Docker Hub (https://hub.docker.com/_/postgres)
  └── PostgreSQL 公式チームが Dockerfile を書いてメンテナンス
      └── タグ: 17-alpine
          └── この image を我々がダウンロードして使わせてもらっている
```

**我々は Postgres のインストール方法を知らなくても、image を使えば Postgres が動く。** それが公式 image の強力さ。

### 4.5 もし Dockerfile が必要になるとしたら

このプロジェクトで Dockerfile を書く必要が出るケース:

#### (A) Postgres に改造を加えたい
```dockerfile
FROM postgres:17-alpine
RUN apk add --no-cache postgis  # 地理空間拡張を追加
```

#### (B) 初期化時にシードデータを流したい
```dockerfile
FROM postgres:17-alpine
COPY init-data.sql /docker-entrypoint-initdb.d/
```

#### (C) Go アプリを Docker 化することにした
- 将来デプロイするなら本番用のマルチステージ Dockerfile を書く
- ローカル開発も Docker 化するなら開発用 Dockerfile を書く

**現時点ではこれらのニーズがない** → Dockerfile が存在しない。

---

## 5. 意思決定フローチャート(サービスごとに判定)

```
このサービスをどう動かすか
         │
         ▼
    Docker で動かす必要があるか?
    (インストール手間、環境統一、永続化、リセット容易性 etc.)
         │
    ┌────┴────┐
   Yes         No
    │          │
    │          ▼
    │     ホストで直接実行 (go run 等)
    │     → Docker も Dockerfile も不要
    │
    ▼
    既成 image で要件を満たせるか?
         │
    ┌────┴────┐
   Yes         No
    │          │
    ▼          ▼
 compose.yml  Dockerfile を書く
 に `image:`  (build: で指定)
 だけ書く
 → Dockerfile 不要
```

### このプロジェクトの判定結果

| サービス | Q1. Docker 必要? | Q2. 既成 image で OK? | 結論 |
|----------|-------------------|------------------------|------|
| Go アプリ | **No** | (問わない) | ホスト実行、Dockerfile 不要 |
| Postgres | **Yes** | **Yes**(公式 `postgres:17-alpine`) | `image:` 指定のみ、Dockerfile 不要 |

→ **結果: Dockerfile がプロジェクトに 1 つも存在しない。**

---

## 6. material-hub との対比

| 項目 | `todo-list-backend`(この学習プロジェクト) | `material-hub-backend`(業務プロジェクト) |
|------|------------------|-------------------|
| Go アプリ | ホスト実行 | Docker 化 |
| Dockerfile | **なし** | `Dockerfile.local`(開発用) + `Dockerfile.dev`(デプロイ用) |
| Postgres | 公式 image (`postgres:17-alpine`) | 公式 image (同上) |
| サービス数 | 2(app + db) | 3+(app + db + atlas ...) |
| 共同開発 | 1 人(学習用) | チーム |
| デプロイ予定 | なし | あり(dev / staging / prod クラスタ) |

→ **要件が違えば構成も違う** 典型例。「Dockerfile があるから本格的、ないから手抜き」ではなく、**要件に応じた最適化**。

---

## 7. 将来 Dockerfile を追加する可能性

以下のどれかが発生したら、Dockerfile 追加を検討する:

### (A) デプロイする必要が出た
本番 Linux コンテナで動かすなら、マルチステージ Dockerfile を用意:

```dockerfile
# ---- build stage ----
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

### (B) チーム開発を始める
全員で環境を統一するため、`Dockerfile.local` + compose.yml で bind mount + rerun 構成に寄せる。

### (C) Postgres に拡張機能を入れたくなった
PostGIS、pg_cron、TimescaleDB などを導入するなら、Postgres 公式 image をベースに Dockerfile を書く。

---

## 8. まとめ

### 1 行要約

> **Go アプリはホストで `go run` する方が学習に向き、Postgres は公式 image で要件を満たせるため、このプロジェクトには Dockerfile が 1 つも存在しない。"Dockerfile がないのは手抜き" ではなく、"自前で image を作る必要が一切ない構成" の結果。**

### 覚えておきたい判断軸

| 場面 | Docker 化 | Dockerfile |
|------|-----------|------------|
| ミドルウェア(Postgres / Redis 等) | 個人開発でも積極的に使う | 公式 image で足りれば不要 |
| Go アプリ(学習・個人) | ホスト実行で十分 | 不要 |
| Go アプリ(チーム開発 + デプロイ) | Docker 化する価値あり | 必要(開発用 + デプロイ用の 2 つが典型) |

### メンティー向けのキーフレーズ

> 「**Docker 化するかどうか、Dockerfile を書くかどうか、は別々の判断。** 対象の性質(インストール重さ・状態の有無・サービス連携数)と、プロジェクトの性質(開発者数・デプロイ有無・ツール数)で都度決める」

この視点を持っていれば、どんなプロジェクトでも Docker 構成を自分で判断・説明できるようになる。
