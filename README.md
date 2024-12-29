# urlf: URL formatting utility

[![Go Reference](https://pkg.go.dev/badge/github.com/shibukawa/urlf.svg)](https://pkg.go.dev/github.com/shibukawa/urlf)

* [日本語](https://github.com/shibukawa/urlf/blob/main/README.ja.md)

urlf is a Safe URL formatting library for Go. This library wraps `net.URL` and provides a printf style function that is short and easy to read and escapes safely the values in the URL template to prevent URL injection.

## Installation

```bash
$ go get github.com/shibukawa/urlf
```

## Basic Usage

```go
import (
    "net/http"
    "github.com/shibukawa/urlf"
)

http.Get(urlf.Urlf("https://example.com/api/users/{}/profile", 1000))
```

## URL Template Rules

You can put placeholders in the following locations:

- protocol (`string` or `*string`)
- hostname (`string` or `*string`)
- port (`int` or `*int`)
- path (`string` or `*string`, `int` `*int` or `[]any`)
- query value  (`string` or `*string`, `int`, `*int`)
- query set (`url.Values`)
- fragment (`string` or `*string`)

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

Placeholder can be written only between each delimiter (`://`, `:`, `/`, `?`, `=`, `&`, `#`) and interpolated strings are escaped properly.

### Path Hierarchies

Placeholders for path can accept slice or `/` separated string, and it supports URLs with variable path hierarchies.

```go
areaList := []string{"japan", "tokyo", "shinjuku"};
urlf.Urlf(`https://example.com/menu/{}`, areaList)
// => 'https://example.com/menu/japan/tokyo/shinjuku'

areaStr := "japan/tokyo/shinjuku";
urlf.Urlf`https://example.com/menu/{}`, areaStr)
// => 'https://example.com/menu/japan/tokyo/shinjuku'
```

### nil

If the placeholder value is `nil`, the placeholder and related text (like query key) are removed from the resulting URL. To pass `nil`, the pointer type are acceptable for placeholder variables.

```go
var port     *string
var value1   *string
var value2   *string
var fragment *string

value2 =  &[]string{"value2"}[0]

urlf.Urlf(`https://example.com:{}/api/users?key1={}&key2={}#{}`, port, value1, value2, fragment)
// => 'https://example.com/api/users?key2=value2'
```

This behavior is useful when you want to implement paging query that has default values than `url.Values`.

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

### Query Set

It accepts `url.Values` instance as a query set and merges it with other queries.

```go
searchParams := url.Values{
    "word": []string{"spicy food"},
    "safeSearch": []string{"false"},
    "spicyLevel": []string{"Infinity"},
}
urlf.Urlf(`https://example.com/api/search?{}`, searchParams)
// => 'https://example.com/api/search?word=spicy+food&safeSearch=false'
```

## Advanced Usage

Custom factory function can overwrite the some parts of the URL. It is good for specifies the API host that is from environment variables or credentials that should not be hard-coded in the source code:

- `protocol`
- `hostname`: It can contains `protocol` and/or `port`.
- `port`
- `username` and `password`: It is only available location to define in this library.

```go
apiUrl := urlf.CustomFormatter(urlf.Opt{
    Hostname: os.Getenv("API_SERVER_HOST"),  // https://localhost:8080
    Username: os.Getenv("API_SERVER_USER"),  // user
    Password: os.Getenv("API_SERVER_PASS"),  // pAssw0rd

})

apiUrl(`https://api-server/api/users/{}/profile`, 1000)
// => 'https://user:pAssw0rd@localhost:8080/api/users/1000/profile'
// "https://api-server" is a dummy string that is replaced with customFormatter()'s hostname option.
// You can avoid hard-coding the actual hostname in your project code.
```

## License

Apache-2.0

## References

* TypeScript version: [url-tidy](https://www.npmjs.com/package/url-tidy)