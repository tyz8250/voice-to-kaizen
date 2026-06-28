# voice-to-kaizen v0.1 Issues

## Milestone 0: Project Setup

### Issue #1 READMEにコンセプトを書く

目的: voice-to-kaizen の思想と作る理由を明確にする。

完了条件:

- READMEにプロジェクト名を書く
- サブタイトルを書く
- コンセプトを書く
- 「改善を、個人の善意に依存させない。」を書く
- v0.1で作るものを書く
- v0.1で作らないものを書く
- 実際の業務システムとして運用しているものではないことを書く

### Issue #2 Goプロジェクトを初期化する

目的: Go APIサーバーの土台を作る。

完了条件:

- `go mod init` を実行する
- `cmd/server/main.go` を作る
- `/healthz` が `200 OK` を返す
- READMEに起動方法を書く

確認:

```bash
curl http://localhost:8080/healthz
```

### Issue #3 Docker ComposeでPostgreSQLを起動する

目的: ローカル開発用のPostgreSQLをDocker Composeで起動することができるようにする。このIssueではGoアプリからのDB接続はまだ行わない。

完了条件:

- `docker-compose.yml` を作る
- `.env` を作る
- PostgreSQLコンテナを起動できる
- DB名は `voice_to_kaizen`
- user/password は `.env` で管理する
- `docker compose ps` でpostgresが起動していることを確認できる
- READMEに起動方法を書く

確認:

```bash
docker compose up -d
docker compose ps
```

### Issue #4 GoからPostgreSQLに接続する

目的: APIサーバーからDB接続できるようにする。

完了条件:

- `DATABASE_URL` を環境変数から読み込む
- `pgxpool` でPostgreSQLに接続する
- `/healthz/db` がDB接続確認を返す

確認:

```bash
curl http://localhost:8080/healthz/db
```

## Milestone 1: Database and Auth

### Issue #5 migrationを導入する

目的: DBスキーマをコードで管理する。

完了条件:

- migrationツールを導入する
- `migrations` ディレクトリを作る
- `users`, `kaizen_requests`, `comments`, `status_histories`, `decision_logs` のDDLを作る
- migrate up/down ができる

### Issue #6 seedユーザーを作る

目的: ログイン確認用の初期ユーザーを作る。

完了条件:

- `admin@example.com` のseedユーザーを作る
- `password_hash` を保存する
- role は `admin`
- READMEに初期ログイン情報を書く

### Issue #7 ログインAPIを作る

API: `POST /login`

目的: JWT認証の入口を作る。

完了条件:

- email/password でログインできる
- 成功時にJWTを返す
- 失敗時は401を返す
- `password_hash` と照合する

### Issue #8 JWT認証middlewareを作る

目的: 管理系APIを認証必須にする。

完了条件:

- `Authorization: Bearer <token>` を検証する
- 認証成功時に `user_id`, `role` をcontextへ入れる
- 認証失敗時は401を返す
- `/healthz` と `/login` は認証不要

## Milestone 2: Kaizen Requests

### Issue #9 改善要望投稿APIを作る

API: `POST /kaizen-requests`

目的: 現場の声を改善チケットとして登録できるようにする。

リクエスト例:

```json
{
  "title": "申請受付後の二重入力を減らしたい",
  "description": "Excelと基幹システムの両方に同じ内容を入力しており、転記ミスが発生している。",
  "category": "duplicate_input",
  "impact": 5,
  "urgency": 4,
  "effort": 3
}
```

完了条件:

- JWT必須
- `title`, `description`, `category`, `impact`, `urgency`, `effort` を保存する
- `priority_score` を service 層で計算する
- `requester_id` をログインユーザーにする
- `owner_id` が未設定なら status は `owner_needed`
- 作成日時を保存する

### Issue #10 改善要望一覧APIを作る

API: `GET /kaizen-requests`

目的: 改善要望を一覧できるようにする。

完了条件:

- JWT必須
- 改善要望一覧を取得できる
- `priority_score` の降順で並べる
- `status`, `category`, `owner_id` で絞り込みできる

### Issue #11 改善要望詳細APIを作る

API: `GET /kaizen-requests/{id}`

目的: 改善チケットの詳細を見られるようにする。

完了条件:

- JWT必須
- 指定IDの改善要望を取得できる
- comments と status_histories も取得できる

### Issue #12 優先度スコア計算serviceを作る

目的: 業務ルールをservice層に分離する。

ルール:

```text
priority_score = impact * urgency - effort
```

完了条件:

- `CalculatePriorityScore(impact, urgency, effort int) int` を作る
- `impact`, `urgency`, `effort` は1〜5の範囲にする
- 範囲外ならエラーにする
- テストを書く

テスト名:

```go
func TestCalculatePriorityScore(t *testing.T) {}
func TestRejectsInvalidPriorityInputs(t *testing.T) {}
```

## Milestone 3: Ownership and Action

### Issue #13 担当者割り当てAPIを作る

API: `POST /kaizen-requests/{id}/assign`

目的: 改善要望にownerを設定できるようにする。

リクエスト例:

```json
{
  "owner_id": 2
}
```

完了条件:

- JWT必須
- `owner_id` を設定できる
- status が `owner_needed` の場合、`planned` に変更する
- `status_histories` に履歴を残す

### Issue #14 next_action登録APIを作る

API: `POST /kaizen-requests/{id}/next-action`

目的: 改善要望に次の行動と期限を設定できるようにする。

リクエスト例:

```json
{
  "next_action": "現行Excelと基幹システムの入力項目を洗い出す",
  "due_date": "2026-07-05"
}
```

完了条件:

- JWT必須
- `next_action` を保存する
- `due_date` を保存する
- 空の `next_action` は400にする

### Issue #15 ステータス変更APIを作る

API: `POST /kaizen-requests/{id}/status`

目的: 改善要望の状態を進められるようにする。

リクエスト例:

```json
{
  "status": "in_progress",
  "reason": "入力項目の洗い出しを開始したため"
}
```

完了条件:

- JWT必須
- status を変更できる
- `from_status`, `to_status`, `changed_by`, `reason` を `status_histories` に保存する
- 不正なstatusは400にする

## Milestone 4: Comments and Decisions

### Issue #16 コメント追加APIを作る

API: `POST /kaizen-requests/{id}/comments`

目的: 改善要望にコメントを残せるようにする。

リクエスト例:

```json
{
  "body": "繁忙期に転記ミスが起きやすいので、優先度は高めでよさそうです。"
}
```

完了条件:

- JWT必須
- コメントを追加できる
- `author_id` をログインユーザーにする
- 空コメントは400にする

### Issue #17 decision log追加APIを作る

API: `POST /kaizen-requests/{id}/decisions`

目的: 「やる」「やらない」「保留」などの判断理由を残せるようにする。

リクエスト例:

```json
{
  "decision": "planned",
  "reason": "影響範囲が広く、対応コストも低いため次月対応に入れる"
}
```

完了条件:

- JWT必須
- decision と reason を保存する
- `decided_by` をログインユーザーにする
- 空のreasonは400にする

## Milestone 5: Visibility

### Issue #18 owner未設定一覧APIを作る

API: `GET /kaizen-requests/owner-needed`

目的: 担当者が決まっておらず止まっている改善要望を見える化する。

完了条件:

- JWT必須
- status が `owner_needed` の改善要望だけ取得する
- 作成日時が古い順で並べる

### Issue #19 期限切れ一覧APIを作る

API: `GET /kaizen-requests/overdue`

目的: `next_action` の期限が切れている改善要望を見える化する。

完了条件:

- JWT必須
- `due_date` が今日より前
- status が `done` / `rejected` / `archived` ではない
- `due_date` が古い順で並べる

### Issue #20 完了一覧APIを作る

API: `GET /kaizen-requests/done`

目的: 完了した改善を見える化する。

完了条件:

- JWT必須
- status が `done` の改善要望を取得する
- 更新日時が新しい順で並べる

## Milestone 6: Demo and Finish

### Issue #21 curlデモを書く

目的: READMEを見る人が一連の流れを確認できるようにする。

完了条件:

- `docs/demo.md` を作る
- 以下の流れをcurlで再現できる

```text
1. login
2. 改善要望投稿
3. 一覧取得
4. 担当者割り当て
5. next_action登録
6. status変更
7. コメント追加
8. owner未設定一覧
9. 期限切れ一覧
10. 完了一覧
```

### Issue #22 READMEにデモシナリオを書く

目的: 現場課題の解像度を伝える。

デモケース:

```text
申請受付後、Excelと基幹システムに同じ内容を二重入力している。
繁忙期に転記ミスが発生している。
```

完了条件:

- READMEにデモシナリオを書く
- 改善チケット化された例を書く
- `priority_score` の計算例を書く
- このプロジェクトが実業務利用ではなくポートフォリオであることを書く

### Issue #23 v0.1.0タグを切る

目的: v0.1として完成範囲を固定する。

完了条件:

- READMEにv0.1の完成範囲を書く
- 今後やることを書く
- `v0.1.0` タグを作る

## Recommended Build Order

まずはこの順で進める。

```text
#1 README
#2 /healthz
#3 Docker Compose PostgreSQL
#4 DB接続
#5 migration
#6 seed user
#7 login
#8 JWT
#9 改善要望投稿
#12 優先度スコア計算
```

ここまでで、プロジェクトの心臓が動き出す。

次に進めるもの:

```text
#13 assign
#14 next_action
#15 status変更
#18 owner未設定一覧
#19 期限切れ一覧
```

ここまで行けば、「改善を善意に依存させない」が機能として見える。
