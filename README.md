# voice-to-kaizen

現場の声を、改善チケットに変える業務改善API。

## Subtitle

現場の声を、改善チケットに変える業務改善API

## Concept

業務改善は、声の大きい人や責任感のある人だけで進めるものではありません。

現場には「困っている」「面倒くさい」「なんとかしたい」という声があります。しかし、その声は口頭・メール・Excel・会議メモに散らばり、担当者や次の行動が決まらないまま流れてしまうことがあります。

voice-to-kaizen は、そうした現場の声を改善チケットとして受け取り、影響度・緊急度・対応コスト・担当者・次の行動・期限・履歴を持たせることで、改善を個人の善意に依存させず、仕組みで前に進めることを目指すAPIです。

## Core Message

改善を、個人の善意に依存させない。

## What This Project Is

このプロジェクトは、実際の業務システムとして職場で運用しているものではありません。現職で感じた業務改善上の課題を題材に、要件定義・API設計・DB設計・Goでの実装を学ぶために作成するポートフォリオです。

## v0.1 Scope

### Build

- ユーザー登録またはseedユーザー
- ログインAPI
- 改善要望の投稿
- 改善要望の一覧取得
- 改善要望の詳細取得
- impact / urgency / effort による priority_score 計算
- owner 割り当て
- next_action 登録
- due_date 登録
- status 変更
- コメント追加
- status_history 保存
- owner未設定一覧
- 期限切れ一覧
- READMEとcurlデモ

### Do Not Build in v0.1

- 画面UI
- 通知機能
- メール送信
- AI分類
- kintone連携
- AWSデプロイ
- 複雑な組織階層
- 実際の業務利用を前提にしたセキュリティ完全対応

## Getting Started

### Run API Server

```bash
go run ./cmd/server
```

The server starts on `:8080` by default. Set `PORT` to change it.

### Health Check

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

### Database Health Check

Start PostgreSQL first, then run the API server with `DATABASE_URL`.

```bash
docker compose up -d
DATABASE_URL="postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable" go run ./cmd/server
```

In another terminal:

```bash
curl http://localhost:8080/healthz/db
```

Expected response:

```json
{"status":"ok"}
```

## Docker Compose

前提: Docker Desktopがインストールされていること
Docker version 27.4.0
Docker Compose version v2.31.0-desktop.2

## PostgreSQLの起動

```bash
docker compose up -d
docker compose ps
```

## PostgreSQLの停止

```bash
docker compose down
```

## Database Migration

このプロジェクトでは、golang-migrate をCLIツールとして利用します。

migrationファイルは `migrations` ディレクトリに配置します。
Goアプリの起動時に自動でmigrationは実行しません。
DBの構造変更は、migrateコマンドを手動で実行して反映します。

### Install golang-migrate

macOS:

```bash
brew install golang-migrate
```

```bash
migrate -version
```

環境変数の設定方法とmigrationコマンドの詳細は、[migrations/README.md](migrations/README.md)を参照してください。Issue #5-1では手順の準備まで行い、Issue #5-2/#5-3でテーブル作成とup/down確認を行います。

### Run users migration

PostgreSQLを起動し、Mac側のターミナルへ `DATABASE_URL` を設定します。

```bash
docker compose up -d
export DATABASE_URL='postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable'
```

未適用のmigrationを実行します。

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

`users` テーブルの作成結果を確認します。`docker compose exec` にはコンテナ名ではなくサービス名の `db` を指定します。

```bash
docker compose exec db psql -U voice_user -d voice_to_kaizen -c '\d users'
```

直前のmigrationを1つ戻します。

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
```

テーブルが削除されたことを確認します。

```bash
docker compose exec db psql -U voice_user -d voice_to_kaizen -c '\dt'
```

確認後、開発を続ける場合は再度 `up` を実行して最新状態へ戻します。

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

## Seeds data

開発環境では、ログイン確認用の初期ユーザーをseed SQLで作成しています。

以下のコマンドで実行できます。

```bash
docker compose exec -T db psql -U voice_user -d voice_to_kaizen < seeds/001_admin_user.sql
```

- [Admin user seed](seeds/001_admin_user.sql)

同じseed SQLを複数回実行しても、`ON CONFLICT DO NOTHING` により重複レコードが挿入されません。

### 初期ログインユーザー

- メールアドレス: `admin@example.com`
- パスワード: `admin123`
- ロール: `admin`

## Docs

- [Project brief](docs/project-brief.md)
- [v0.1 issues](docs/issues.md)
- [Issue #5-2 usersテーブル作成で学んだこと](docs/issue-05-2-users-migration-notes.md)
- [Database schema](docs/database.md)
- [Issue #5-4 kaizen_requestsテーブル再設計ログ](docs/issue-05-04.md)

## Related Articles

- [GoでDATABASE_URLがわからなかったので、os.Getenv・環境変数・.envの関係を整理した](https://tyzo8250.hatenablog.jp/entry/2026/06/27/131902)
- [Goでpgxpool.Newがわからなかったので、PostgreSQLの接続プールを整理した](https://tyzo8250.hatenablog.jp/entry/2026/06/28/132141)
