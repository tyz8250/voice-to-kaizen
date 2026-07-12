package main

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func handleLogin(dbpool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// DB接続プールがnilの場合、Panicせずに`503 Service Unavailable`を返す
		if dbpool == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "database unavailable",
			})
			return
		}

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

		err := dbpool.QueryRow(
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

		// パスワードをハッシュ化して比較
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

		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"email":  email,
		})
	}
}
