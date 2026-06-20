# voice-to-kaizen Project Brief

## Subtitle

現場の声を、改善チケットに変える業務改善API

## What This Project Is

voice-to-kaizen は、現場の「困っている」「面倒くさい」「なんとかしたい」という声を、改善チケットに変える業務改善APIです。

このプロジェクトは、実際の業務システムとして職場で運用しているものではありません。現職で感じた業務改善上の課題を題材に、要件定義・API設計・DB設計・Goでの実装を学ぶために作成するポートフォリオです。

## Technical Stack

- Go
- PostgreSQL
- Docker Compose
- REST API
- JWT認証
- migration
- READMEとcurlデモを重視

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

## Implementation Policy

- 1 Issue = 1機能で小さく進める
- 完了条件を満たす最小実装にする
- 既存構成を大きく変えない
- handler / service / repository / model を意識する
- 業務ルールは service 層に置く
- DB操作は repository 層に置く
- curlで動作確認できるようにする
- READMEにプロジェクトの思想と実行方法を書く

## Database Design v0.1

### users

- id
- name
- email
- password_hash
- role
- created_at
- updated_at

`role` はまず `admin` と `member` にする。

### kaizen_requests

- id
- title
- description
- category
- status
- impact
- urgency
- effort
- priority_score
- requester_id
- owner_id
- next_action
- due_date
- created_at
- updated_at

### comments

- id
- request_id
- author_id
- body
- created_at

### status_histories

- id
- request_id
- from_status
- to_status
- changed_by
- reason
- created_at

### decision_logs

- id
- request_id
- decided_by
- decision
- reason
- created_at

## Statuses

v0.1で最低限使うstatus:

- captured
- owner_needed
- planned
- in_progress
- done
- rejected

将来候補:

- triage_needed
- triaged
- waiting
- archived

## Priority Score

```text
priority_score = impact * urgency - effort
```

- impact: 影響度。どれくらい多くの人・業務に効くか
- urgency: 緊急度。どれくらい早く対応すべきか
- effort: 対応コスト。どれくらい大変か

Example:

```text
impact = 5
urgency = 4
effort = 3
priority_score = 17
```

## API v0.1

- `POST /login`
- `POST /kaizen-requests`
- `GET /kaizen-requests`
- `GET /kaizen-requests/{id}`
- `POST /kaizen-requests/{id}/assign`
- `POST /kaizen-requests/{id}/next-action`
- `POST /kaizen-requests/{id}/status`
- `POST /kaizen-requests/{id}/comments`
- `POST /kaizen-requests/{id}/decisions`
- `GET /kaizen-requests/owner-needed`
- `GET /kaizen-requests/overdue`
- `GET /kaizen-requests/done`

## Demo Data

```json
{
  "title": "申請受付後の二重入力を減らしたい",
  "description": "新システム導入後も、確認用Excelと基幹システムの両方に同じ内容を入力している。繁忙期に入力漏れや転記ミスが発生している。",
  "category": "duplicate_input",
  "impact": 5,
  "urgency": 4,
  "effort": 3
}
```

```text
priority_score = 5 * 4 - 3 = 17
```

Next action:

```text
現行Excelと基幹システムの入力項目を洗い出す
```

