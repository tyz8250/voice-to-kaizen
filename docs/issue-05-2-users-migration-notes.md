# Issue #5-2 usersテーブル作成で学んだこと

## このIssueの目的

Issue #5-2では、`users` テーブルを作成・削除するためのmigrationファイルを用意します。

今回の対象:

- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`

今回の対象外:

- `kaizen_requests` など、`users` 以外のテーブル
- Goコードの変更
- Goアプリ起動時のmigration自動実行
- migrate up/downの正式な動作確認（Issue #5-3で実施）

## 1. migrate upを実行する前にPostgreSQLを起動する

`migrate up` は、接続先のPostgreSQLへmigrationファイルのSQLを送って実行するコマンドです。そのため、先にPostgreSQLが起動している必要があります。

```bash
docker compose up -d
docker compose ps
```

確認の流れ:

```text
Docker ComposeでPostgreSQLを起動する
        ↓
DATABASE_URLを環境変数へ設定する
        ↓
migrate upでSQLを適用する
        ↓
psqlで結果を確認する
```

## 2. DATABASE_URLを明示的に設定する

次のエラーは、`DATABASE_URL` が空のときに発生します。

```text
error: failed to parse scheme from database URL: URL cannot be empty
```

コマンドに `"$DATABASE_URL"` と書いても、環境変数が未設定なら空文字として渡されます。先に接続情報を設定します。

```bash
export DATABASE_URL='postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable'
```

設定されたことは、次のコマンドで確認できます。

```bash
echo "$DATABASE_URL"
```

## 3. `.env` とシェルの環境変数は別のもの

`.env` に値を書いただけでは、現在のターミナルで環境変数として使えるとは限りません。

```text
.envに値が書かれている
    ≠
ターミナルで環境変数が設定されている
```

Docker Composeはプロジェクト直下の `.env` を変数展開に利用します。一方、Macにインストールした `migrate` CLIへ接続情報を渡す場合は、`export DATABASE_URL=...` などでシェルの環境変数を設定する必要があります。

## 4. `docker compose exec`にはサービス名を使う

このプロジェクトのComposeサービス名は `db` です。

```text
NAME       : voice-to-kaizen-db
SERVICE    : db
```

`docker compose exec` で指定するのは、コンテナ名の `voice-to-kaizen-db` ではなくサービス名の `db` です。

```bash
docker compose exec db psql -U voice_user -d voice_to_kaizen
```

サービス名は次のコマンドで確認できます。

```bash
docker compose ps
```

## 5. DB名とユーザー名は `.env` の値から決まる

`docker-compose.yml` では、PostgreSQLの初期設定に `.env` の値を使用しています。

```yaml
environment:
  POSTGRES_DB: ${POSTGRES_DB}
  POSTGRES_USER: ${POSTGRES_USER}
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
```

対応関係:

| `.env` | 用途 |
| --- | --- |
| `POSTGRES_DB` | `psql` の `-d` とDATABASE_URLのDB名 |
| `POSTGRES_USER` | `psql` の `-U` とDATABASE_URLのユーザー名 |
| `POSTGRES_PASSWORD` | DATABASE_URLのパスワード |

`psql -U postgres` で次のエラーが出た場合、`postgres` というroleが作成されていないことを表します。

```text
FATAL: role "postgres" does not exist
```

このプロジェクトでは、`.env` の `POSTGRES_USER` に設定したユーザーを使用します。

## 6. 実行場所によって接続先hostが変わる

MacへHomebrewでインストールした `migrate` CLIからDocker上のPostgreSQLへ接続する場合、公開ポートを経由するためhostは `localhost:5432` です。

```text
MacからDocker上のDBへ接続する
→ localhost:5432
```

Goアプリなど別のDocker ComposeサービスからPostgreSQLへ接続する場合は、Composeネットワーク内のサービス名を使用します。

```text
DockerコンテナからDBコンテナへ接続する
→ db:5432
```

## 7. `migrate` と `psql` の役割は異なる

```text
migrate up
→ migrationファイルのSQLをDBへ適用する

psql
→ DBへ接続して、適用結果を確認する
```

つまり、migrationでテーブルを作り、`psql` でテーブルが作られたかを確認します。

テーブル一覧:

```psql
\dt
```

`users` テーブルの構造:

```psql
\d users
```

## 8. migrationファイルがあれば `.gitkeep` は不要

Gitは空のディレクトリを管理しません。ただし、このプロジェクトでは `migrations/README.md` があるため、Issue #5-1の時点から `migrations` ディレクトリをGitで管理できます。

Issue #5-2ではup/down SQLも追加されるため、`.gitkeep` を追加する必要はありません。

## 9. Issueを小さく保つ

Issue #5-2では `users` テーブルのmigrationファイル作成に集中します。

次の作業は別Issueで行います。

- migrate up/downの確認: Issue #5-3
- `kaizen_requests` テーブル作成: Issue #5-4
- その他の関連テーブル作成: Issue #5-5

## 今回のまとめ

SQLの内容だけでなく、次の要素がどのようにつながるかを理解することが重要でした。

- Docker Compose
- `.env`
- `DATABASE_URL`
- Composeのサービス名
- PostgreSQLのDB名とユーザー名
- `migrate` CLI
- `psql`

特に重要な点:

```text
migrate upはPostgreSQLへSQLを適用する
PostgreSQLは先に起動しておく
接続情報はDATABASE_URLでmigrate CLIへ渡す
docker compose execにはサービス名のdbを使う
DB名とユーザー名は.envの設定に合わせる
```

## Issue #5-2と#5-3の関係

今回の作業では、migrationファイルの作成だけでなく、実際の `migrate up`、`psql`による確認、`migrate down` まで行いました。

そのため、作業内容としては次の2つを含んでいます。

- Issue #5-2: `users` テーブルのup/down用DDLを作る
- Issue #5-3: migrate up/downとテーブルの作成・削除を確認する

レビューで既存設計との不一致が見つかったため、`users` テーブルには `password_hash`、`role`、`updated_at` も含めました。修正後のmigrationでup/downを再確認します。
