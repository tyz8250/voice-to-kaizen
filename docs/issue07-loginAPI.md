# Issue #7 ログインAPIを作る

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

### 実装予定

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

## 今後の予定

- #7-5 bcryptでパスワードを照合する
- #7-6 JWTを発行する
- #7-7 成功・失敗パターンをテストする
- #7-8 email・passwordの空欄チェックを追加する
