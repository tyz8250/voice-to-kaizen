# Issue #4 GoからPostgreSQLに接続する

## 目的

このIssueでは、APIサーバーからPostgreSQLに接続できるかを確認します。

まだテーブル作成やmigrationは行いません。まずはGoアプリケーションがPostgreSQLへ接続できる状態を作ることだけを目的にします。

```mermaid
flowchart LR
ブラウザ / curl
   |
   | GET /healthz/db
   v
Go APIサーバー
   |
   | dbpool.Ping()
   v
PostgreSQL
   |
   | つながった
   v
{"status":"ok"}
```

## 今回やること

- `DATABASE_URL` を環境変数から読み込む
- `pgxpool` でPostgreSQLに接続する
- `/healthz/db` でDB接続確認を返す

## 今回やらないこと

- テーブル作成
- migration導入
- seedデータ作成
- CRUD API実装

## 実装の考え方

### 1. DB接続情報はコードに直接書かない

PostgreSQLへ接続するには、DB名、ユーザー名、パスワード、接続先ホスト、ポート番号が必要です。

これらをGoコードに直接書いてしまうと、開発環境や本番環境で値を変えにくくなります。そのため、`DATABASE_URL` という環境変数から読み込む形にします。

例:

```text
postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable
```

このURLには次の情報が含まれています。

- `voice_user`: PostgreSQLのユーザー名
- `voice_password`: PostgreSQLのパスワード
- `localhost`: 接続先ホスト
- `5432`: PostgreSQLのポート番号
- `voice_to_kaizen`: DB名
- `sslmode=disable`: ローカル開発ではSSL接続を使わない設定

### 2. GoからPostgreSQLへ接続するためにpgxpoolを使う

GoからPostgreSQLに接続するために、`github.com/jackc/pgx/v5/pgxpool` を使います。

`pgxpool` はPostgreSQL接続を管理するための仕組みです。APIサーバーではリクエストのたびにDBへ接続する可能性があるため、接続を毎回作り直すより、接続poolとして管理する方が扱いやすくなります。

今回のIssueではDB操作は行わないので、接続poolを作り、接続確認として `Ping` だけを使います。

### 3. `/healthz/db` ではPingだけを行う

`/healthz/db` はDB接続確認用のエンドポイントです。

このエンドポイントでは、テーブルを読んだり、データを書き込んだりしません。`dbpool.Ping(ctx)` を実行して、PostgreSQLに接続できるかだけを確認します。

接続できた場合:

```json
{"status":"ok"}
```

接続できない場合:

```json
{
  "status": "error",
  "error": "database unavailable"
}
```

### 4. `/healthz` と `/healthz/db` の役割を分ける

`/healthz` はAPIサーバー自体が起動しているかを確認するためのエンドポイントです。

一方で、`/healthz/db` はAPIサーバーからPostgreSQLへ接続できるかを確認するためのエンドポイントです。

この2つは確認したい対象が違います。

- `/healthz`: APIサーバーが動いているか
- `/healthz/db`: APIサーバーからDBへ接続できるか

そのため、`DATABASE_URL` が未設定でも `/healthz` は動きます。DB接続情報がない場合は、`/healthz/db` だけ `503 Service Unavailable` を返します。

## 変更したファイル

### `cmd/server/main.go`

主な変更内容:

- `DATABASE_URL` を `os.Getenv` で読み込む
- `pgxpool.New` でPostgreSQL接続poolを作る
- `GET /healthz/db` を追加する
- `/healthz/db` で `dbpool.Ping(ctx)` を実行する
- DB接続成功時は `{"status":"ok"}` を返す
- `DATABASE_URL` 未設定時は `/healthz/db` だけ `503` を返す

### `.env.example`

`DATABASE_URL` の例を追加しました。

```text
DATABASE_URL=postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable
```

### `go.mod` / `go.sum`

PostgreSQL接続用に `github.com/jackc/pgx/v5` を追加しました。

## 動作確認

PostgreSQLを起動します。

```bash
docker compose up -d
```

`DATABASE_URL` を指定してAPIサーバーを起動します。

```bash
DATABASE_URL="postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable" go run ./cmd/server
```

別のターミナルからDB接続確認を行います。

```bash
curl http://localhost:8080/healthz/db
```

成功すると、次のレスポンスが返ります。

```json
{"status":"ok"}
```

## ここまでで確認できること

このIssueが完了すると、Go APIサーバーからPostgreSQLへ接続できることが確認できます。

ただし、まだテーブルは作成していません。次のIssue以降でmigrationを導入し、必要なテーブルを作成していきます。
