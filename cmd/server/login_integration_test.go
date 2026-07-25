package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ログインAPIの正常系を確認するテスト
func TestHandleLoginReturnsJWTForValidCredentials(t *testing.T) {
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
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`,
		email,
		string(passwordHash),
		"user",
	).Scan(&userID)

	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	// 2. emailとpasswordを入れたリクエストを作る
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
}
