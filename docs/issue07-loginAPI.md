# Issue #7 ログインAPIを作る

## 進捗

- #7-1〜#7-6 完了
- ログイン成功時に、`user_id`と`role`を含むJWTを返すところまで実装済み
- JWTの署名方式はHS256、有効期限は発行から24時間
- 次の作業は#7-7（成功・失敗パターンのテスト）

## #7-1 ログインAPIの仕様を決める

目的：実装前に、入力・出力・失敗パターンを決める。

### 入力

- email
- password

### 出力

- JWT token

### 失敗パターン

- emailがDBに存在しない
- passwordが間違っている
- emailとpasswordが空欄
- リクエストJSONの形式が不正

### レスポンスの例

```http
POST /login
```

Request:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Success Response:

```json
{
  "token": "JWT_TOKEN"
}
```

Failed Response:

```http
401 Unauthorized
```

```json
{
  "error": "invalid email or password"
}
```

---

## #7-2 login request用のstructを作る

目的：JSONのリクエストをGoで受け取れる形にする。

### 完了条件

- `LoginRequest` structを作成する
- `email`、`password`フィールドを持つ
- JSONタグを付与する

```go
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
```

### 疑問

- structをどのフォルダ・ファイルに配置するか
- JSONからGoのstructへ変換される仕組み
- JSONタグがどのように使われるか

---

## #7-3 `/login` handlerの骨組みを作る

目的：まだDB照合は行わず、HTTPリクエストを受け取れる状態にする。

### 完了条件

- `POST /login`を受けられる
- POST以外は`405 Method Not Allowed`を返す
- JSONをdecodeできる
- JSONが不正な場合は`400 Bad Request`を返す
- 仮のレスポンスとして`200 OK`を返す

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// POST以外は405
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JSONをLoginRequestへdecode
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// 仮の成功レスポンス
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"email":  req.Email,
	})
}
```

### 学んだこと

- `json.NewDecoder(r.Body).Decode(&req)`で、リクエストボディのJSONをGoのstructへ変換できる
- `&req`は、decodeした値を書き込む先のアドレスを渡している
- HTTPメソッドがPOST以外の場合は、`Allow`ヘッダーに許可するメソッドを設定する

---

## #7-4 emailでユーザーをDBから検索する

目的：ログインリクエストで受け取ったemailを使って、`users`テーブルから対象ユーザーを取得する。

### 完了条件

- `handleLogin`がDB接続を受け取れる
- emailを条件に`users`テーブルを検索できる
- `id`、`email`、`password_hash`、`role`を取得できる
- ユーザーが見つからない場合は`401 Unauthorized`を返す
- DB検索時にHTTPリクエストのContextを渡す

### handlerの変更

DB接続を利用するため、`handleLogin`が`*pgxpool.Pool`を受け取り、`http.HandlerFunc`を返す形に変更する。

```go
func handleLogin(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handlerの処理
	}
}
```

ルーティング側では、DB接続を渡してhandlerを登録する。

```go
http.HandleFunc("/login", handleLogin(db))
```

### DB検索処理

```go
func handleLogin(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// POST以外は405
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// JSONをLoginRequestへdecode
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		var userID int
		var email string
		var passwordHash string
		var role string

		err := db.QueryRow(
			r.Context(),
			`
			SELECT id, email, password_hash, role
			FROM users
			WHERE email = $1
			`,
			req.Email,
		).Scan(
			&userID,
			&email,
			&passwordHash,
			&role,
		)

		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid email or password",
			})
			return
		}

		// 現時点では仮の成功レスポンス
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"email":  email,
		})
	}
}
```

### 学んだこと

#### `r.Context()`

```go
r.Context()
```

HTTPリクエストに紐づいたContextを取得する。

利用者が通信を切断した場合や、リクエストがキャンセルされた場合、その情報をDB処理にも引き継げる。

```text
HTTPリクエスト
    ↓
r.Context()
    ↓
DB検索
```

HTTPハンドラー内のDB処理では、基本的に`context.Background()`ではなく`r.Context()`を使用する。

#### `$1`

```sql
WHERE email = $1
```

`$1`は、SQLへ安全に値を渡すためのプレースホルダー。

```go
db.QueryRow(
	r.Context(),
	`SELECT id FROM users WHERE email = $1`,
	req.Email,
)
```

この場合の対応関係は次のとおり。

```text
$1 ← req.Email
```

値が複数ある場合は、`$2`、`$3`と番号を増やす。

```sql
WHERE email = $1 AND role = $2
```

文字列を直接SQLへ結合せず、プレースホルダーを使うことで、SQLインジェクションを防止できる。

### コミット

```bash
go fmt ./...
go test ./...

git status
git add .
git commit -m "feat: find user by email in login handler"
```

---

## #7-5 bcryptでパスワードを照合する

目的：リクエストで受け取ったパスワードと、DBに保存されている`password_hash`を照合する。

### 完了条件

- `bcrypt.CompareHashAndPassword`を使用する
- DBの`password_hash`と入力されたpasswordを照合する
- パスワードが間違っている場合は`401 Unauthorized`を返す
- パスワードが正しい場合は次のJWT発行処理へ進める

### 実装

```go
err = bcrypt.CompareHashAndPassword(
	[]byte(passwordHash),
	[]byte(req.Password),
)
if err != nil {
	writeJSON(w, http.StatusUnauthorized, map[string]string{
		"error": "invalid email or password",
	})
	return
}
```

### ポイント

入力されたパスワードを再度ハッシュ化して比較するのではなく、bcryptの照合関数を使用する。

```text
DBに保存されたpassword_hash
入力された生のpassword
        ↓
bcrypt.CompareHashAndPassword
```

---

## #7-6 JWTを発行する

目的：ログインに成功したユーザーへJWTを発行し、レスポンスとして返す。

**ステータス：完了**

### 完了条件

- [x] JWT生成用のライブラリを導入する
- [x] JWTへ`user_id`と`role`を含める
- [x] JWTに発行日時と有効期限を設定する
- [x] 環境変数`JWT_SECRET`を使って署名する
- [x] ログイン成功時にJWTを返す
- [x] JWT生成に失敗した場合は`500 Internal Server Error`を返す
- [x] curlでJWTが返ることを確認する

---

### #7-6-1 JWTライブラリを追加する

JWTの生成には、`github.com/golang-jwt/jwt/v5`を使用する。

プロジェクトのルートディレクトリで、以下を実行する。

```bash
go get github.com/golang-jwt/jwt/v5
```

実行後、`go.mod`と`go.sum`へ依存関係が追加される。

---

### #7-6-2 JWTの秘密鍵を用意する

JWTの署名に使用する秘密鍵を、環境変数`JWT_SECRET`として設定する。

`.env`へ追加する。

```env
JWT_SECRET=十分に長いランダムな文字列
```

秘密鍵はGitへ登録しない。

`.env.example`には、実際の秘密鍵ではなく設定項目だけを書く。

```env
JWT_SECRET=your-jwt-secret
```

ランダムな文字列を生成する場合は、以下のコマンドを利用できる。

```bash
openssl rand -base64 32
```

`.env`を自動で読み込む仕組みがない場合は、サーバー起動前に環境変数を設定する。

```bash
export JWT_SECRET="生成した秘密鍵"
```

設定を確認する。

```bash
echo $JWT_SECRET
```

秘密鍵そのものは、ログやGitHubへ公開しない。

---

### #7-6-3 JWTに含める情報を定義する

`cmd/server/jwt.go`を作成する。

JWTには、以下の情報を含める。

- `user_id`：ログインしたユーザーのID
- `role`：ユーザーの権限
- `iat`：JWTを発行した日時
- `exp`：JWTの有効期限

```go
package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type LoginClaim struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
```

`jwt.RegisteredClaims`を埋め込むことで、JWTで標準的に使用される`iat`や`exp`を設定できる。

実装では、型名を`LoginClaim`としている。

---

### #7-6-4 JWTを生成する関数を作る

`jwt.go`へ、JWTを生成する関数を追加する。

```go
func generateJWT(userID int, role string, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	now := time.Now()

	claims := LoginClaim{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
}
```

### 処理の流れ

```text
ユーザーIDとroleを受け取る
↓
JWTへ入れるclaimsを作成する
↓
署名方式HS256を指定する
↓
JWT_SECRETで署名する
↓
JWT文字列を返す
```

### `generateJWT`の引数

```go
func generateJWT(userID int, role string, secret string)
```

- `userID`：DBから取得したユーザーID
- `role`：DBから取得したユーザー権限
- `secret`：JWTの署名に使用する秘密鍵

### 戻り値

```go
(string, error)
```

JWTの生成に成功した場合は、JWT文字列を返す。

失敗した場合は、空文字とエラーを返す。

---

### #7-6-5 ログイン成功時にJWTを生成する

bcryptによるパスワード照合が成功した後に、JWTを生成する。

`login.go`で環境変数を読み取るため、`os`をimportする。

```go
import "os"
```

bcrypt照合処理の後に、以下を追加する。

```go
jwtSecret := os.Getenv("JWT_SECRET")

token, err := generateJWT(userID, role, jwtSecret)
if err != nil {
	log.Printf("failed to generate JWT: %v", err)

	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "internal server error",
	})
	return
}
```

JWT生成に失敗した場合、利用者へ秘密鍵などの詳細は返さない。

詳細なエラーはサーバー側のログへ記録し、レスポンスには一般的なエラーメッセージを返す。

---

### #7-6-6 成功レスポンスでJWTを返す

これまで使用していた仮の成功レスポンスを変更する。

変更前：

```go
writeJSON(w, http.StatusOK, map[string]string{
	"status": "ok",
	"email":  email,
})
```

変更後：

```go
writeJSON(w, http.StatusOK, map[string]string{
	"token": token,
})
```

ログインに成功すると、以下のようなJSONが返る。

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### ログイン処理全体の流れ

```text
POST /login
↓
JSONをLoginRequestへdecode
↓
emailでusersテーブルを検索
↓
password_hashを取得
↓
bcryptでパスワードを照合
↓
user_idとroleをJWTへ入れる
↓
JWT_SECRETで署名
↓
JWTをレスポンスとして返す
```

---

### #7-6-7 curlでJWTの発行を確認する

最初にDockerを起動する。

```bash
docker compose up -d
```

DB接続を確認する。

```bash
curl -i http://localhost:8080/healthz/db
```

`200 OK`が返ることを確認する。

次に、Mac側のターミナルで`DATABASE_URL`と`JWT_SECRET`を設定してサーバーを起動する。このプロジェクトは`.env`を自動では読み込まないため、環境変数を明示的に設定する。

```bash
export DATABASE_URL='postgres://voice_user:voice_password@localhost:5432/voice_to_kaizen?sslmode=disable'
export JWT_SECRET="$(openssl rand -base64 32)"
go run ./cmd/server
```

別のターミナルから、正しいemailとpasswordを送信する。

```bash
curl -i \
  -X POST \
  http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }'
```

期待するレスポンス：

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "token": "JWT_TOKEN"
}
```

間違ったパスワードも確認する。

```bash
curl -i \
  -X POST \
  http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "wrong-password"
  }'
```

期待するレスポンス：

```http
HTTP/1.1 401 Unauthorized
```

```json
{
  "error": "invalid email or password"
}
```

---

### 学んだこと

#### JWTはユーザー情報を署名付きで表現する

JWTには、ログインしたユーザーのIDやroleなどを含められる。

ただし、JWTの中身は暗号化されているとは限らず、利用者から確認できる。

そのため、以下のような秘密情報はJWTへ入れない。

- パスワード
- password_hash
- JWT_SECRET
- 個人情報や機密情報

#### JWT_SECRETはコードへ直接書かない

以下のように秘密鍵をコードへ直接記述すると、GitHubなどへ公開される可能性がある。

```go
// やらない
secret := "my-secret-key"
```

環境変数から取得する。

```go
secret := os.Getenv("JWT_SECRET")
```

#### JWTには有効期限を設定する

JWTを無期限にすると、流出した場合に長期間悪用される可能性がある。

今回は、有効期限を24時間に設定する。

```go
ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour))
```

#### JWT発行はパスワード照合成功後に行う

JWTは、emailとpasswordが正しいことを確認した後にだけ発行する。

```text
DB検索成功
↓
パスワード照合成功
↓
JWT発行
```

認証に失敗した利用者へJWTを発行してはいけない。

---

### 実装後の確認

```bash
go fmt ./...
go test ./...
git status
```

問題がなければcommitする。

```bash
git add .
git commit -m "feat: issue JWT on successful login"
```

### #7-6 完了時の実装結果

- `cmd/server/jwt.go`に`LoginClaim`と`generateJWT`を追加した
- `github.com/golang-jwt/jwt/v5`を利用してHS256で署名した
- JWTへ`user_id`、`role`、`iat`、`exp`を格納した
- `cmd/server/login.go`でbcrypt照合成功後にJWTを生成し、`token`として返した
- `JWT_SECRET`が空の場合やJWT生成に失敗した場合は、詳細をレスポンスへ出さず`500 Internal Server Error`を返した
- `go test ./...`が成功することを確認した

### 参考
- [https://jwt.io/ja/introduction](https://jwt.io/ja/introduction)
- [https://pkg.go.dev/github.com/golang-jwt/jwt/v5](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [https://datatracker.ietf.org/doc/html/rfc7519](https://datatracker.ietf.org/doc/html/rfc7519)
- [https://www.rfc-editor.org/info/rfc8725](https://www.rfc-editor.org/info/rfc8725)

## 今後の予定

- #7-7 成功・失敗パターンをテストする
- #7-8 email・passwordの空欄チェックを追加する
