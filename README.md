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

## Docs

- [Project brief](docs/project-brief.md)
- [v0.1 issues](docs/issues.md)
