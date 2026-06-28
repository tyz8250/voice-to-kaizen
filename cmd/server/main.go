package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// コンテキストを作成
	ctx := context.Background()

	dbpool, err := openDBPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/healthz/db", handleDBHealthz(dbpool))

	// 環境変数からポート番号を取得し、設定がない場合は8080をデフォルト値とする。
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 指定されたポートで待機する
	addr := ":" + port
	// ログ出力
	log.Printf("voice-to-kaizen API listening on %s", addr)
	// サーバー起動
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// openDBPool はPostgreSQLの接続プールを作成します
func openDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL is not set; /healthz/db will return 503")
		return nil, nil
	}

	return pgxpool.New(ctx, databaseURL)
}

// handleHealthz はヘルスチェック用エンドポイントを処理します
// GET /healthzにアクセスされたら、status: okを返却します
// それ以外のメソッドは405 Method Not Allowedを返却します
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		// メソッド not allowed を返却
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// レスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	// ステータスコードを設定
	w.WriteHeader(http.StatusOK)

	// レスポンスをJSON形式でエンコード
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("write health response: %v", err)
	}
}

// handleDBHealthz はDBヘルスチェック用エンドポイントを処理します
// GET /healthz/dbにアクセスされたら、status: okを返却します
// それ以外のメソッドは405 Method Not Allowedを返却します
func handleDBHealthz(dbpool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// DB接続プールがnilの場合
		if dbpool == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "error",
				"error":  "DATABASE_URL is not set",
			})
			return
		}

		// DB接続確認用のコンテキストを作成。2秒でタイムアウト
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// DB接続を確認
		if err := dbpool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "error",
				"error":  "database unavailable",
			})
			log.Printf("database health check failed: %v", err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}
