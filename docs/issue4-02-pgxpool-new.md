はてなブログに投稿するためのメモ

### はじめに

前回の記事で、`DATABASE_URL`、`os.Getenv`、環境変数、`.env`の関係を整理しました。

その中で、Goのコードが .env ファイルを直接読んでいるわけではなく、実行環境に設定された環境変数を `os.Getenv` で取得しているのだと理解しました。

ただ、関数全体を見ると、まだ読めない部分が残っています。

```go
// openDBPool はPostgreSQLの接続プールを作成します
func openDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL is not set; /healthz/db will return 503")
		return nil, nil
	}

	return pgxpool.New(ctx, databaseURL)
}
```

特に、最後の行の
```go
return pgxpool.New(ctx, databaseURL)
```

が何を意味しているのかがわかりません。

今回は、`pgxpool.New`は何を行なっているのか整理していこうと思います。

話の広がりを防ぐため、この記事では`context`については触れません。

### pgxとpgxpoolとは何か

まずは、`pgx`と`pgxpool`とは何か調べていきます。

自分のコードを読んでいると以下のように `pgxpool`を`import` していました。

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)
```

ここで、そもそも何を`import`しているのかを理解する必要があると感じました。

調べてみると、`pgx`はGoからPostgreSQLを扱うためのライブラリで、`pgxpool` はその接続プールを扱うためのパッケージだとわかりました。

ここで出てきたのが、「単一接続」と「接続プール」という言葉です。

単一接続は、CLIツールや一度だけDBに接続して処理を終えるような場合には、シンプルで扱いやすそうです。

一方で、Web APIのように複数のリクエストが同時に来る可能性があるアプリでは、毎回DB接続を作るよりも、接続プールで管理する方が向いているようです。

今回作っている `voice-to-kaizen`はAPIサーバーです。

今後、複数のAPIエンドポイントからデータベースにアクセスする予定もあります。

そのため、単一接続ではなく `pgxpool`を使うのは自然だと理解しました。

全て後付けではありますが、少なくとも今回の方針は大きく間違っていなさそうです。

### pgxpool.Newは何をしているのか

では`pgxpool.New`が書いてある行を見ていきます。

```go
return pgxpool.New(ctx, databaseURL)
```

databaseURL がPostgreSQLへの接続情報であることは、前回の記事で整理しました。

そして、`pgxpool`が接続プールを扱うためのパッケージであることもわかりました。

では、`pgxpool.New(ctx, databaseURL)`は何をしているのでしょうか。

ここで一度、公式ドキュメントを確認します。

>Package pgxpool is a concurrency-safe connection pool for pgx.

>The primary way of creating a pool is with pgxpool.New:

日本語に訳すと：

>pgxpoolはpgxの並行安全な接続プールです。

>プールを作成する主な方法はpgxpool.Newです。

と説明されています。

つまり `pgxpool.New`は、その名前のとおり、新しい接続プールを作成するための関数だと理解しました。

今回のコードでは、以下のように `ctx`と `databaseURL`を渡しています。

```go
return pgxpool.New(ctx, databaseURL)
```

私の今の理解では、この1行は、

「databaseURL に書かれた接続情報を使って、PostgreSQLへの接続プールを作成して返す」

という処理です。

もう少し分解すると、以下のように読めそうです。

* pgxpool.New
  * 新しい接続プールを作成する関数
* ctx
  * 接続処理のキャンセルやタイムアウトに関係するもの
* databaseURL
  * PostgreSQLへの接続情報

ここでまた `ctx`が出てきました。

`context`についても気になりますが、ここまで深掘りすると話が広がりそうです。

そのため今回は、`pgxpool.New(ctx, databaseURL)`は「PostgreSQLへの接続プールを作る処理」だと理解するところまでにします。

`context`については、次回以降で整理したいと思います。

### まとめ：pgxpool.Newは接続プールを作る処理だった

今回は、`pgx` と `pgxpool`、そして `pgxpool.New(ctx, databaseURL)` が何をしているのかを整理しました。

`pgx` はGoからPostgreSQLを扱うためのライブラリで、`pgxpool` は接続プールを扱うためのパッケージです。

今回作っている `voice-to-kaizen` はAPIサーバーなので、複数のリクエストからDBにアクセスする可能性があります。

そのため、単一接続ではなく、接続プールを使う `pgxpool` を選ぶのは自然だと理解しました。

また、`pgxpool.New(ctx, databaseURL)` は、`databaseURL` に書かれた接続情報を使って、PostgreSQLへの接続プールを作成する処理だとわかりました。

ただし、ここでも `ctx` が出てきます。

`context` についてはまだ理解が浅いため、次回以降で整理していきたいと思います。

## Issue #4でまだ理解が浅いこと

- `context.Context` が何をしているのか
- `context.WithTimeout` がなぜ必要なのか
- Docker Compose内で `localhost` と `db` の違いがどう生まれるのか
- GoコンテナからPostgreSQLコンテナへどう名前解決しているのか

これらはIssue #5以降でも関係するため、後続で改めて整理する。
