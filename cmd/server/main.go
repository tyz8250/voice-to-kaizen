package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)

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
