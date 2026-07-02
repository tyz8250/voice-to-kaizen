# Database Migrations

このディレクトリには、データベースの構造変更を管理するmigrationファイルを配置します。

このプロジェクトでは、[`golang-migrate`](https://github.com/golang-migrate/migrate)をCLIツールとして使用します。Goアプリの起動時にはmigrationを自動実行せず、必要なタイミングでCLIから手動実行します。

## Directory Role

今後、次のようなSQLファイルを追加します。

```text
000001_create_users_table.up.sql
000001_create_users_table.down.sql
```

- `*.up.sql`: データベースの構造を新しい状態へ進めるSQL
- `*.down.sql`: `up`で行った変更を前の状態へ戻すSQL

現在はまだmigrationファイルを作成していません。このREADMEがあることで、空のmigrationファイルがなくても `migrations` ディレクトリをGitで管理できます。

## Prerequisites

migrationを実行する前に、PostgreSQLを起動します。

```bash
docker compose up -d
```

次に、接続先を `DATABASE_URL` 環境変数へ設定します。

```bash
export DATABASE_URL='postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable'
```

`golang-migrate` CLIは `.env` を自動では読み込みません。上記のようにシェルで環境変数を設定してからコマンドを実行します。接続情報の例は `.env.example` でも確認できます。

この接続情報はローカル開発用の例です。実際のパスワードをREADMEやGit管理対象のファイルへ書かないでください。

## Migration Commands

以下のコマンドは、migrationファイルを作成するIssue #5-2以降で使用します。Issue #5-1では手順を記載するだけで、まだ `up` や `down` は実行しません。

未適用のmigrationを実行します。

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

直前のmigrationを1つ戻します。

```bash
migrate -path migrations -database "$DATABASE_URL" down 1
```

現在のmigration versionを確認します。

```bash
migrate -path migrations -database "$DATABASE_URL" version
```

## Issue #5-1 Scope

Issue #5-1で行うこと:

- `golang-migrate` CLIをインストールする
- `migrations` ディレクトリを作成する
- migrationコマンドの使い方を文書化する

Issue #5-1では行わないこと:

- migration用SQLファイルの作成
- テーブルの作成や削除
- migrate up/downの動作確認

実際のテーブル作成とup/downの動作確認は、後続Issueで行います。
