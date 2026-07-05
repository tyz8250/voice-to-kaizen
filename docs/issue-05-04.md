# Issue #5-4 kaizen_requestsテーブル再設計ログ

## 目的

Issue #5-4では、改善要望を保存する `kaizen_requests` テーブルを追加します。最初の設計で不足していたカラムとstatus値を見直し、Issue #5-4自体をやり直す形でversion 2を再設計しました。

このテーブルは、改善要望を出した依頼者を `requester_id`、担当者を `owner_id` として保持し、それぞれ `users.id` を参照します。

テーブル間の関係は[データベースER図](database.md)で確認できます。

## 実施したこと

### 1. usersテーブルとの関係をER図で整理した

- `kaizen_requests.requester_id` は `users.id` を参照する
- `kaizen_requests.owner_id` は `users.id` を参照する
- 1人のユーザーは複数の改善要望を依頼できる
- 1人のユーザーは複数の改善要望の担当者になれる

### 2. 外部キーを理解した

FKはForeign Keyの略で、日本語では外部キーと呼びます。外部キーは、別テーブルの主キーを参照するためのカラムです。

```text
users
id | name
1  | 山田
2  | 佐藤

kaizen_requests
id | requester_id | owner_id | title
1  | 1            | 2        | 申請フォームを改善したい
```

この例では、山田さんが依頼者、佐藤さんが担当者です。ユーザー名を `kaizen_requests` に重複保存せず、`users.id` で関連付けます。

外部キー制約があることで、`users` テーブルに存在しないIDを `requester_id` や `owner_id` に登録できなくなります。

### 3. migrationファイルを作成した

```text
migrations/000002_create_kaizen_requests_table.up.sql
migrations/000002_create_kaizen_requests_table.down.sql
```

- up: `kaizen_requests` テーブルを作成する
- down: `kaizen_requests` テーブルだけを削除する

`users` テーブルはversion 1で作成しているため、version 2のdownでは削除しません。

このプロジェクトはまだ開発初期で、`kaizen_requests` に保存済みデータがありません。そのため、後続の修正migrationは追加せず、version 2の定義を完成形へ書き直しました。

### 4. migrate upを実行した

```bash
docker compose up -d
export DATABASE_URL='postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable'
migrate -path migrations -database "$DATABASE_URL" up
```

テーブル一覧は、Composeサービス名の `db` を指定して確認します。

```bash
docker compose exec db psql -U voice_user -d voice_to_kaizen -c '\dt'
```

## 詰まったこと

### migration version 2なのにテーブルが表示されなかった

migration実行中に失敗すると、`schema_migrations` にversionとdirty状態が記録されます。

`schema_migrations` は、`golang-migrate` が現在のmigration versionを管理するために使用するテーブルです。

```text
migrate up
    ↓
version 2の実行途中で失敗
    ↓
schema_migrations: version=2, dirty=true
    ↓
以降のup/downが停止する
```

### Dirty database version 2になった

表示されたエラー:

```text
error: Dirty database version 2. Fix and force version.
```

今回の復旧では、DBの実際の状態を確認したうえで管理versionを1へ戻しました。

```bash
migrate -path migrations -database "$DATABASE_URL" force 1
```

`force` はmigration SQLを実行しません。テーブルを作成・削除する処理ではなく、`schema_migrations` のversionとdirty状態を書き換える操作です。

そのため、原因となったSQLやDBの状態を確認せずに `force` を実行してはいけません。中途半端に作成されたDBオブジェクトがある場合は、先に手動で状態を整える必要があります。

### FOREIGN付近でSQL構文エラーが出た

表示されたエラー:

```text
migration failed: syntax error at or near "FOREIGN"
```

原因は `FOREIGN KEY` 自体ではなく、その直前にある `updated_at` 定義のカンマ漏れでした。

誤り:

```sql
updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
CONSTRAINT fk_requester
  FOREIGN KEY (requester_id)
  REFERENCES users(id),
```

修正後:

```sql
updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT fk_requester
  FOREIGN KEY (requester_id)
  REFERENCES users(id),
```

SQLエラーで示された位置は、必ずしも原因そのものの位置とは限りません。直前のカラム定義やカンマも確認する必要があります。

### SQL修正後に再実行した

DBの状態を確認してversion 1へ戻し、SQLを修正してから再度upを実行しました。

```bash
migrate -path migrations -database "$DATABASE_URL" force 1
migrate -path migrations -database "$DATABASE_URL" up
```

レビュー時点では、DBは次の状態です。

```text
version: 2
dirty: false
kaizen_requests: 作成済み
```

## 今回理解したこと

- `kaizen_requests` は `users` に依存する
- `requester_id` は改善要望を出した人を表す
- `owner_id` は改善要望の担当者を表す
- 外部キーは存在しない `users.id` の登録を防ぐ
- migration失敗時はdirty状態になる
- dirty状態ではup/downが停止する
- `force` はSQLを実行せず、migration管理情報だけを書き換える
- SQLエラーの本当の原因が、表示位置の直前にあることもある
- `CREATE TABLE` 内のカラム定義と制約定義はカンマで区切る

## レビューを受けた再設計

### owner_idをNULL許可にした

担当者未設定の改善要望を扱えるように、`owner_id` はNULLを許可しました。値が入っている場合は、外部キー制約によって実在する `users.id` だけを登録できます。

### 必要なカラムをversion 2へ追加した

プロジェクト設計に合わせて、次のカラムをversion 2へ追加しました。

- `category`
- `impact`
- `urgency`
- `effort`
- `priority_score`
- `next_action`
- `due_date`

### status値をv0.1設計へ揃えた

当初の `open / in_progress / closed` をやめ、v0.1の設計に合わせて次の値を使用します。

- `captured`
- `owner_needed`
- `planned`
- `in_progress`
- `done`
- `rejected`

プロジェクト設計を正として、version 2のCHECK制約を統一しました。

### up/down確認

再設計したversion 2で、次の内容を確認しました。

1. `down 1` で `kaizen_requests` だけが削除される
2. `users` テーブルは残る
3. 再度upすると `kaizen_requests` が作成される

確認後は再度upし、開発DBをversion 2の最新状態へ戻しました。

```text
最終migration version: 2
users: 存在する
kaizen_requests: 存在する
```
