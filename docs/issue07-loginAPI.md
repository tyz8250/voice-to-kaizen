# Issue #7 ログインAPIを作る

## #7-1 ログインAPIの仕様を決める

目的：実装前に、入力・出力・失敗パターンを決める。

### 入力：

- email
- password

### 出力：

- JWT token

### 失敗パターン：

- emailがDBに存在しない
- passwordが間違っている
- emailとpasswordが空欄

### レスポンスの例

```json
POST /login

Request:
{
  "email": "user@example.com",
  "password": "password123"
}

Success Response:
{
  "token": "JWT_TOKEN"
}

Failed Response:
401 Unauthorized
{
  "error": "Invalid email or password"
}
```

## 7-2 login request 用の struct を作る

目的：JSONのリクエストをGoで受け取れる形にする。

完了条件：

- LoginRequest struct を作成する
- email, password フィールドを持つ
- JSONタグを付与する

```go
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

疑問：
- struct をどのフォルダ・ファイルに配置するか
- Json-->Go の変換はどんな原理なのか（優先ではない）

## 7-3 /login の handler の骨組みを作る

目的：まだDB照合しなくていいので、HTTPとして受け取れる状態にする。

完了条件：

* POST /login を受けられる
* POST以外は405
* JSONをdecodeできる
* 仮で200を返す

例：

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
}
```
