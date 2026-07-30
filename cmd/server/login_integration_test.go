package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ログインAPIの正常系を確認するテスト
func TestHandleLoginReturnsJWTForValidCredentials(t *testing.T) {
	jwtSecret := "test-jwt-secret"
	t.Setenv("JWT_SECRET", jwtSecret)

	// テスト用DBを作製
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")

	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set") // 環境変数が設定されていない場合はスキップ（ログインAPIが正常か判断できないため）
	}

	ctx := context.Background() // DB接続用のコンテキスト（何も設定されていない）
	dbpool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to create db pool: %v", err)
	}
	t.Cleanup(dbpool.Close)

	// テスト用DBに接続できることを確認する
	if err := dbpool.Ping(ctx); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	password := "test-password"
	name := "Login Test User"
	role := "member"

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf("generate password has: %v", err)
	}

	email := fmt.Sprintf(
		"login-test-%d@example.com",
		time.Now().UnixNano(),
	)

	var userID int

	// 1. テスト用ユーザーを準備する
	err = dbpool.QueryRow(
		ctx,
		`
		INSERT INTO users (name,email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		name,
		email,
		string(passwordHash),
		role,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	// テスト終了時に作成したユーザを削除する
	t.Cleanup(func() {
		_, err := dbpool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("delete test user: %v", err)
		}
	})

	// emailとpasswordを入れたリクエストを作る
	requestBody := fmt.Sprintf(
		`{"email": %q, "password": %q}`,
		email,
		password,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(requestBody),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// 3. handleLoginを実行する
	handler := handleLogin(dbpool)
	handler(recorder, req)

	// 4. 200 OKか確認する
	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d, body = %s",
			recorder.Code,
			http.StatusOK,
			recorder.Body.String(),
		)
	}

	// 5. レスポンスにtokenがあるか確認する
	var response struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if response.Token == "" {
		t.Fatal("token is empty")
	}

	claims := &LoginClaim{}

	token, err := jwt.ParseWithClaims(
		response.Token, // ログインAPIから返されたJWT
		claims,         // JWTの中身を入れる箱
		// 秘密鍵を返す無名関数
		func(token *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		},
	)

	// 解析・署名検証が成功したか確認
	if err != nil {
		t.Fatalf("parse JWT: %v", err)
	}

	// tokenが有効か確認する
	if !token.Valid {
		t.Fatal("token is invalid")
	}

	// claims.UserID-->JWTから取り出したID
	// userID（テスト用のユーザーID）
	if claims.UserID != userID {
		t.Errorf(
			"user_id = %d, want %d",
			claims.UserID,
			userID,
		)
	}

	// roleの確認
	if claims.Role != role {
		t.Errorf(
			"role = %q, want %q",
			claims.Role,
			role,
		)
	}

}

// emailは正しい、PWが間違っている場合のテスト
/*
登録済みemail
+
間違ったpassword
↓
401 Unauthorized
↓
errorが期待どおり
↓
tokenは空
*/
func TestHandleLoginRejectWrongPassword(t *testing.T) {

	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")

	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	dbpool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to create db pool: %v", err)
	}
	t.Cleanup(dbpool.Close)

	if err := dbpool.Ping(ctx); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	password := "correct-password"
	wrongPassword := "wrong-password"
	name := "wrong Password Test User"
	role := "member"
	email := fmt.Sprintf(
		"wrong-password-test-%d@example.com",
		time.Now().UnixNano(),
	)

	// 正しいパスワードをハッシュ化する
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	var userID int

	err = dbpool.QueryRow(
		ctx,
		`
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		name,
		email,
		string(passwordHash),
		role,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, err := dbpool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("delete test user: %v", err)
		}
	})

	// ログインリクエストでは誤ったPWを入力する
	requestBody := fmt.Sprintf(
		`{"email": %q, "password": %q}`,
		email,
		wrongPassword,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(requestBody),
	)
	req.Header.Set("Content-Type", "application/json")

	// handlerのレスポンスを記録する
	recorder := httptest.NewRecorder()

	handler := handleLogin(dbpool)
	handler(recorder, req)

	// handler実行後にレスポンスを確認する
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status code = %d, want %d, body = %s",
			recorder.Code,
			http.StatusUnauthorized,
			recorder.Body.String(),
		)
	}

	var response struct {
		Error string `json:"error"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	wantError := "invalid email or password"

	if response.Error != wantError {
		t.Errorf("error = %q, want %q", response.Error, wantError)
	}

	if response.Token != "" {
		t.Errorf("token = %q, want empty string", response.Token)
	}

}

func TestHandleLoginRejectsUnknownEmail(t *testing.T) {
	// Arrange：テストDBへ接続する
	// テスト用DBを作製
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")

	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set") // 環境変数が設定されていない場合はスキップ（ログインAPIが正常か判断できないため）
	}

	ctx := context.Background() // DB接続用のコンテキスト（何も設定されていない）
	dbpool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to create db pool: %v", err)
	}
	defer dbpool.Close()

	// Arrange：DBに存在しないemailを準備する
	email := fmt.Sprintf("unknown-email-%d@example.com",
		time.Now().UnixNano(),
	)

	password := "any-password"

	// Arrange：ログインリクエストを作る
	requestBody := fmt.Sprintf(
		`{"email": %q, "password": %q}`,
		email,
		password,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(requestBody),
	)
	req.Header.Set("Content-Type", "application/json")

	// Act：handlerを実行する

	recoder := httptest.NewRecorder()
	handler := handleLogin(dbpool)
	handler(recoder, req)

	// Assert：401・error・tokenを確認する
	if recoder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status code = %d, want %d, body = %s",
			recoder.Code,
			http.StatusUnauthorized,
			recoder.Body.String(),
		)
	}

	var response struct {
		Error string `json:"error"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(recoder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	wantError := "invalid email or password"

	if response.Error != wantError {
		t.Errorf("error = %q, want %q", response.Error, wantError)
	}

	if response.Token != "" {
		t.Errorf("token = %q, want empty string", response.Token)
	}
}

func TestHandleLoginRejectsMissingJWTSecret(t *testing.T) {
	// Arrange：テストDBへ接続する
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")

	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set") // 環境変数が設定されていない場合はスキップ（ログインAPIが正常か判断できないため）
	}

	ctx := context.Background() // DB接続用のコンテキスト（何も設定されていない）
	dbpool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to create db pool: %v", err)
	}
	defer dbpool.Close()

	// Arrange：JWT_SECRETを空にする
	t.Setenv("JWT_SECRET", "")

	// Arrange：正しいパスワードをハッシュ化する
	password := "correct-password"

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	// Arrange：テストユーザーをDBへ登録する
	name := "Missing JWT Secret Test User"
	role := "member"

	email := fmt.Sprintf(
		"missing-jwt-secret-%d@example.com",
		time.Now().UnixNano(),
	)

	var userID int

	err = dbpool.QueryRow(
		ctx,
		`
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		name,
		email,
		string(passwordHash),
		role,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	// Cleanup：ユーザー削除を予約する
	t.Cleanup(func() {
		_, err := dbpool.Exec(
			context.Background(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)
		if err != nil {
			t.Errorf("delete test user: %v", err)
		}
	})

	// Arrange：正しいemailとpasswordのリクエストを作る
	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(fmt.Sprintf(
			`{"email": %q, "password": %q}`,
			email,
			password,
		)),
	)
	req.Header.Set("Content-Type", "application/json")

	// Act：レスポンスを受け取る箱を作る
	recorder := httptest.NewRecorder()

	// Act：ログインhandlerを作って実行する
	handler := handleLogin(dbpool)
	handler(recorder, req)
	// Assert：500を確認する
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status code = %d, want %d, body = %s",
			recorder.Code,
			http.StatusInternalServerError,
			recorder.Body.String(),
		)
	}

	// Assert：JSONをresponseへデコードする
	var response struct {
		Error string `json:"error"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Assert：errorを確認する
	wantError := "failed to generate token"

	if response.Error != wantError {
		t.Errorf(
			"error = %q, want %q",
			response.Error,
			wantError,
		)
	}

	// Assert：tokenが空か確認する
	if response.Token != "" {
		t.Errorf(
			"token = %q, want empty string",
			response.Token,
		)
	}
}
