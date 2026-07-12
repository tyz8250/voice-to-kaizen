# Issue #7 ログインAPI実装での振り返り

## 起きたこと

ログインAPIに、メールアドレスによるユーザー検索とbcryptによるパスワード照合を追加した。

実装中、以下の問題が発生した。

- ログイン用のDB検索処理を、誤って`handleDBHealthz`の中に記述した
- `req is undefined`エラーが発生した
- Dockerと`DATABASE_URL`の準備をせずにログインAPIを実行した
- `dbpool`が`nil`の状態で`QueryRow`を呼び出し、panicが発生した
- 正しいパスワードを入力しているつもりだったが、実際には入力したパスワードが間違っていた

## 原因

### 1. 似た形のhandlerへ処理を貼り付けた

`handleDBHealthz`と`handleLogin`は、どちらもDB接続プールを受け取り、`http.HandlerFunc`を返す形になっている。

```go
func handleDBHealthz(dbpool *pgxpool.Pool) http.HandlerFunc
```

```go
func handleLogin(dbpool *pgxpool.Pool) http.HandlerFunc
```

関数の形だけを見て、処理の役割を十分に確認せずコードを追加したため、ログイン処理をDBヘルスチェック用handlerへ書いてしまった。

`handleDBHealthz`の役割はDBへの接続確認であり、ユーザー検索やパスワード照合を行う場所ではない。

### 2. 実行前の環境確認が不足していた

Dockerを起動せず、`DATABASE_URL`も設定されていない状態でサーバーを起動した。

その結果、`dbpool`が`nil`になったまま、以下の処理を実行した。

```go
dbpool.QueryRow(...)
```

これにより、nil pointer dereferenceのpanicが発生し、curl側では次のエラーとなった。

```text
curl: (52) Empty reply from server
```

### 3. エラーの原因をコードやDBに限定して考えた

正しいパスワードでログインできなかったため、以下を疑った。

- migrationが実行されていない
- seedが登録されていない
- bcryptハッシュが壊れている
- SQLやパスワード照合処理が間違っている

実際には、curlで入力していたパスワード自体が間違っていた。

コードや環境だけでなく、テストに使用している入力値も確認する必要がある。

## 学んだこと

### handlerごとの責務を確認する

DBを使用する処理であっても、どのhandlerへ書いてもよいわけではない。

```text
handleDBHealthz
└── DBへ接続できるか確認する

handleLogin
├── JSONを受け取る
├── emailでユーザーを検索する
├── password_hashを取得する
├── bcryptでパスワードを照合する
└── JWTを発行する
```

コードを追加する前に、現在編集している関数名と、その関数の役割を確認する。

### コンパイルエラーは処理の置き場所を見直す手掛かりになる

今回の`req is undefined`は、単に変数が不足していたのではなく、ログインリクエストが存在しない別のhandlerへ処理を書いていることを示していた。

変数が見つからない場合は、変数の宣言漏れだけでなく、処理を記述しているスコープや関数が正しいかも確認する。

### 実行前に依存関係を確認する

ログインAPIはPostgreSQLへ接続するため、テスト前に以下を確認する。

```bash
docker compose up -d
docker compose ps
echo $DATABASE_URL
curl -i http://localhost:8080/healthz/db
```

`/healthz/db`が`200 OK`になってから、ログインAPIをテストする。

### nilの可能性がある値は防御する

`dbpool`が`nil`の場合、panicさせず`503 Service Unavailable`を返す。

```go
if dbpool == nil {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{
		"error": "database unavailable",
	})
	return
}
```

### DB検索エラーをすべて401にしない

ユーザーが存在しない場合と、DB接続やSQLに問題がある場合は分けて扱う。

```go
if err != nil {
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid email or password",
		})
		return
	}

	log.Printf("login user query failed: %v", err)

	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "internal server error",
	})
	return
}
```

これにより、認証失敗とシステム障害を区別できる。

### テスト入力も疑う

期待した結果にならない場合は、コードやDBだけでなく、以下も確認する。

- emailのスペル
- passwordの入力内容
- JSONのキー名
- curlで送信している値
- 使用しているDBや環境変数

## 次回の確認手順

ログインAPIを確認するときは、次の順番で進める。

1. Dockerを起動する
2. `DATABASE_URL`を確認する
3. `/healthz/db`でDB接続を確認する
4. DBに対象ユーザーが存在するか確認する
5. 間違ったパスワードで`401`を確認する
6. 存在しないemailで`401`を確認する
7. 正しいemailとpasswordで`200`を確認する
8. サーバー側のログを確認する
9. `go fmt ./...`と`go test ./...`を実行する

## 今回のまとめ

今回の問題は、Goやbcryptそのものの理解不足だけが原因ではなかった。

- コードを記述する関数の選択
- Dockerや環境変数の準備
- エラーログの読み方
- テスト入力の確認

など、実装以外の確認も重要だと分かった。

エラーが発生したときは、すぐにコードを変更するのではなく、次の順番で切り分ける。

```text
入力値
↓
HTTPリクエスト
↓
handlerの役割とスコープ
↓
環境変数
↓
DB接続
↓
DB内のデータ
↓
bcryptなどの処理
```

一つずつ確認することで、不要な修正を減らし、原因を早く特定できる。
