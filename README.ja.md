# urlf: URLフォーマットユーティリティ

[![Go Reference](https://pkg.go.dev/badge/github.com/shibukawa/urlf.svg)](https://pkg.go.dev/github.com/shibukawa/urlf)

* [English](https://github.com/shibukawa/urlf/blob/main/README.md)

urlfは安全なURLを生成するためのGo用のライブラリです。このライブラリは`net.URL`をラップし、短く読みやすいprintfスタイルの関数を提供します。URLテンプレート内の値を適切にエスケープしてURLインジェクションを防ぎます。

## インストール

```bash
$ go get github.com/shibukawa/urlf
```

## 基本の利用方法

Node.jsやBunの場合のインポートパスは以下の通りです。

```go
import (
    "net/http"
    "github.com/shibukawa/urlf"
)

http.Get(urlf.Urlf("https://example.com/api/users/{}/profile", 1000))
```

## URLテンプレートルール

次の場所にプレースホルダーを置けます。

- プロトコル (`string` もしくは `*string`)
- ホスト名 (`string` もしくは `*string`)
- ポート (`int` もしくは `*int`)
- パス (`string` もしくは `*string`, `int` `*int` or `[]any`)
- クエリーの値 (`string` もしくは `*string`, `int`, `*int`)
- クエリーセット (`url.Values`)
- フラグメント(`string` もしくは `*string`)

```go
protocol    := "https"
hostname    := "example.com"
port        := 8080
path        := "api/users"
queryValue  := "value"
querySet, _ := url.ParseQuery("key1=value1&key2=value2")
fragment    := "fragment"
url.Urlf(`{}://{}:{}/{}?queryKey={}&{}#{}`, protocol, hostname, port, queryValue, querySet, fragment)
```

プレースホルダはそれぞれの区切り記号（`://`、`:`、`/`、`?`、`=`、`&`、`#`）の間にしか書けず、展開された文字列は適切にエスケープされます。

### パス階層

パスのプレースホルダには配列や/区切りの文字列を設定でき、階層が可変のURLにも対応します。

```go
areaList := []string{"japan", "tokyo", "shinjuku"};
urlf.Urlf(`https://example.com/menu/{}`, areaList)
// => 'https://example.com/menu/japan/tokyo/shinjuku'

areaStr := "japan/tokyo/shinjuku";
urlf.Urlf`https://example.com/menu/{}`, areaStr)
// => 'https://example.com/menu/japan/tokyo/shinjuku'
```

### nil

`nil`をプレースホルダーに指定すると、クエリーのキーなどその関連項目ごと消去されて出力されます。 `nil`を渡せるように、プレースホルダーに設定する変数にはポインタ型も使えるようになっています。

```ts
const port = null;
const value1 = null;
const value2 = "value2";
const fragment = null;

url`https://example.com:${port}/api/users?key1=${value1}&key2=${value2}#${fragment}`
// => 'https://example.com/api/users?key2=value2'
```

ページングのクエリーなどの場合は`URLSearchParams`を使うよりもコンパクトに書けます。

```go
var port     *string
var value1   *string
var value2   *string
var fragment *string

value2 =  &[]string{"value2"}[0]

urlf.Urlf(`https://example.com:{}/api/users?key1={}&key2={}#{}`, port, value1, value2, fragment)
// => 'https://example.com/api/users?key2=value2'
```

この動作要素数が可変でデフォルト動作を持つような検索のページングのクエリーを`url.Values`を駆使して組み立てるのよりもシンプルに書けます。

```go
var word    *string
var page    *int
var perPage *int    // use default
var limit   *int    // use default

word = []string{}"spicy food"}[0]
page = []int{10}[0]

urlf.Urlf(`https://example.com/api/search?word={}&page={}&perPage={}&limit={}`, word, page, perPage, limit)
// => 'https://example.com/api/search?word=spicy+food&page=10'
```

### クエリーセット

クエリーを`url.Values`インスタンスでまとめて設定してマージさせることも可能です。

```go
searchParams := url.Values{
    "word": []string{"spicy food"},
    "safeSearch": []string{"false"},
    "spicyLevel": []string{"Infinity"},
}
urlf.Urlf(`https://example.com/api/search?{}`, searchParams)
// => 'https://example.com/api/search?word=spicy+food&safeSearch=false'
```

## より高度な使用方法

カスタムのファクトリー関数を使い、URLの一部を定義して上書きできます。環境変数経由で設定するAPIのホスト名や、ソースコードにハードコードすべきではないクレデンシャル情報を設定するのに便利です。

- `protocol`
- `hostname`: プロトコルやポートも指定可能
- `port`
- `username`、`password`: このライブラリではこの場所でしか設定できません。

```go
apiUrl := urlf.CustomFormatter(urlf.Opt{
    Hostname: os.Getenv("API_SERVER_HOST"),  // https://localhost:8080
    Username: os.Getenv("API_SERVER_USER"),  // user
    Password: os.Getenv("API_SERVER_PASS"),  // pAssw0rd

})

apiUrl(`https://api-server/api/users/{}/profile`, 1000)
// => 'https://user:pAssw0rd@localhost:8080/api/users/1000/profile'
// "https://api-server" はダミーの文字列で、customFormatter()のホスト名オプションで置き換わる
// プロジェクトコードに実際のホスト名をハードコードすることを避けることができます。
```

## License

Apache-2.0

## References

* TypeScript版: [url-tidy](https://www.npmjs.com/package/url-tidy)
